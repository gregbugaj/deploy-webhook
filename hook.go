package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

// ServerConfig configuration
type ServerConfig struct {
	Addr string
}

// HookHandler handles incomming request
func HookHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request recived : %v", r.URL.Path)
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Test " + message
	w.Write([]byte(message))
}

// StartHTTPServer starts http server
func StartHTTPServer(config *ServerConfig) *http.Server {
	server := &http.Server{Addr: config.Addr, Handler: nil}
	http.HandleFunc("/", HookHandler)

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
	// now close the server gracefully ("shutdown")
	// timeout could be given with a proper context
	// (in real world you shouldn't use TODO()).
	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Fatal("HTTP Shutdown Error - ", err)
	}

	log.Printf("main: HTTP server stoped")
}
