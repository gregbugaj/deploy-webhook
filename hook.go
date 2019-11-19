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
)

// ServerConfig configuration
type ServerConfig struct {
	Addr string
}

// DeploymentMessage returned to the user
type DeploymentMessage struct {
	Status string `json:"status"`
	Message string	`json:"message"` 
}

// HookHandler handles incomming request
func HookHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request recived : %v", r.URL.Path)

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	// Read body
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
	name  := "" // repository name
	ref  := "" // The full git ref that was pushed. Example: refs/heads/master.
	commit := ""  // The SHA of the most recent commit on ref after the push.

	// determine source
	if strings.Index(data, "api.github.com") > -1 {
		log.Printf("Github payload recived")
		val, err := url.ParseQuery(data)
		if err!=nil {
			log.Fatal(err)
		}
		payloadJson := val.Get("payload")
		var pushEvent PushEventPayloadGithub
		if err := json.Unmarshal([]byte(payloadJson), &pushEvent); err != nil {
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

	// Execute deployment script
	cmd := exec.Command("./deploy.sh", name, ref, commit)
	log.Printf("Running deployment and waiting for it to finish...")
	payload, err := cmd.Output()

	if err != nil {
		log.Printf("Command finished with error: %v", err)
		status = "error"
	}

	log.Printf("Command finished with \n------------------------------\n %s \n------------------------------", payload)
	bytes, err := json.Marshal(DeploymentMessage {
		Status : status,
		Message : string(payload),
	})
	
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Payload > %s", string(bytes))
	w.Write(bytes)
}

// StartHTTPServer starts http server
func StartHTTPServer(config *ServerConfig) *http.Server {
	server := &http.Server{Addr: config.Addr, Handler: nil}
	http.HandleFunc("/deploy-hook", HookHandler)

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
