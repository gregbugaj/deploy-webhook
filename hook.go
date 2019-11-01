package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"	
	"os"
	"os/signal"
	"os/exec"
	"io/ioutil"
)

// ServerConfig configuration
type ServerConfig struct {
	Addr string
}

// DeplomentMessage returned to the user
type DeplomentMessage struct {
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

	log.Printf("body: %s", body)

	status := "success";
	// Execute deployment script via ansible
	cmd := exec.Command("./deploy.sh")
	log.Printf("Running deployment and waiting for it to finish...")
	payload, err := cmd.Output();

	if err != nil {
		log.Printf("Command finished with error: %v", err)
		status = "error"	
	}

	log.Printf("Command finished with \n------------------------------\n %s \n------------------------------", payload)
	bytes, err := json.Marshal(DeplomentMessage {
		Status : status,
		Message : string(payload),
	})
	
	if err != nil {
		log.Fatal(err)
		panic(err)
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

	// sigal capture
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
