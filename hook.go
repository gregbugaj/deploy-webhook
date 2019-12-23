package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"
)

// startup up time
var startTime string

// ServerConfig configuration
type ServerConfig struct {
	Addr string
}

// DeploymentMessage returned to the user
type DeploymentMessage struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Duration int64  `json:"duration"`
}

// Metric info shows that total number of deployment and last stats for that deployment
type Metric struct {
	Hits     int    `json:"hits"`     // Number of hits
	Commit   string `json:"commit"`   // Commit deployed
	Ref      string `json:"ref"`      // Ref deployed
	Time     string `json:"time"`     // Deployment time
	Duration int64  `json:"duration"` // Duration of the deployment
}

// metrics
var metrics map[string]Metric

// StatusHandler displays status of the service
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received : %v", r.URL.Path)
	msg := "Webhook Service : " + startTime
	w.Write([]byte(msg))
}

// MetricsHandler displays metrics information
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received : %v", r.URL.Path)
	bytes, err := json.Marshal(metrics)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Payload > %s", string(bytes))
	w.Write(bytes)
}

// HookHandler handles incomming request
func HookHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received : %v", r.URL.Path)

	if r.Method != "POST" {
		http.Error(w, "Only POST method is supported", 400)
		return
	}

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	data, err := url.QueryUnescape(string(body))

	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("body: %s", data)
	status := "success"
	name := ""   // repository name
	ref := ""    // The full git ref that was pushed. Example: refs/heads/master.
	commit := "" // The SHA of the most recent commit on ref after the push.

	// determine source

	headers := r.Header
	agent := headers["User-Agent"][0]
	contentType := headers["Content-Type"][0]

	for k, v := range headers {
		fmt.Printf("Header [%s] = %s\n", k, v)
	}

	log.Printf("User-Agent : %s", agent)
	log.Printf("Content-Type : %s", contentType)

	// Data can be POSTed in two methods
	// handling content-types [application/json, application/x-www-form-urlencoded]
	if strings.Index(agent, "GitHub-Hookshot") > -1 || strings.Index(data, "api.github.com") > -1 {
		log.Printf("Github payload received")
		var payloadJSON string

		if contentType == "application/json" {
			payloadJSON = string(body)
		} else { //x-www-form-urlencoded
			val, err := url.ParseQuery(data)
			if err != nil {
				log.Fatal(err)
			}

			payloadJSON = val.Get("payload")
		}

		var pushEvent PushEventPayloadGithub
		if err := json.Unmarshal([]byte(payloadJSON), &pushEvent); err != nil {
			log.Printf("Error decoding body %v", err)
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d", e.Offset)
			}
			log.Fatal(err)
		}

		name = pushEvent.Repository.Name
		ref = pushEvent.Ref
		commit = pushEvent.After
	}

	log.Printf("name : %s", name)
	log.Printf("ref : %s", ref)
	log.Printf("commit : %s", commit)

	metric := metrics[name]
	metric.Hits++
	metric.Commit = commit
	metric.Ref = ref

	tic := nowAsUnixMilliseconds()
	// Execute deployment script
	cmd := exec.Command("./deploy.sh", name, ref, commit)
	log.Printf("Running deployment and waiting for it to finish...")
	payload, err := cmd.Output()

	if err != nil {
		log.Printf("Command finished with error: %v", err)
		status = "error"
	}

	toc := nowAsUnixMilliseconds()
	duration := toc - tic
	metric.Duration = duration
	metric.Time = time.Now().Format("Jan _2 15:04:05")
	metric.Commit = commit
	metric.Ref = ref
	// https://github.com/golang/go/issues/3117
	metrics[name] = metric

	log.Printf("Command finished with \n------------------------------\n %s \n------------------------------", payload)
	bytes, err := json.Marshal(DeploymentMessage{
		Status:   status,
		Message:  string(payload),
		Duration: duration,
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Payload > %s", string(bytes))
	w.Write(bytes)
}

func nowAsUnixMilliseconds() int64 {
	return time.Now().Round(time.Millisecond).UnixNano() / 1e6
}

// StartHTTPServer starts http server and registers handlers
// Exposed endpoints
// /deploy
func StartHTTPServer(config *ServerConfig) *http.Server {
	server := &http.Server{Addr: config.Addr, Handler: nil}
	http.HandleFunc("/", StatusHandler)
	http.HandleFunc("/deploy", HookHandler)
	http.HandleFunc("/metrics", MetricsHandler)

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("HTTP Server Error - ", err)
			panic(err)
		}
	}()

	return server
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "host:port")
		return
	}

	startTime = time.Now().Format("Jan _2 15:04:05")
	metrics = make(map[string]Metric)

	port := os.Args[1]
	log.Printf("main: Server starting port# %v", port)

	conf := ServerConfig{Addr: port} // *":8080"
	srv := StartHTTPServer(&conf)

	// signal capture
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	// Waiting for SIGINT pkill -2 (user presses ctrl-c)
	log.Printf("main: Server ready to accept connections")
	<-done

	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Fatal("HTTP Shutdown Error - ", err)
	}

	log.Printf("main: HTTP server stoped")
}
