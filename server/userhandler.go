package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

const (
	stringType = 0
	intType    = 1
	fileType   = 2
)

type Message struct {
	TaskID   int    `json:"taskID"`
	DataType int    `json:"dataType"`
	Data     []byte `json:"data"`
}

type AuthMessage struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func startUIServer(clearDB bool) {
	// Main function responsible for handling client connections

	if clearDB {
		setupDatabase()
	}
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	//
	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel, args Message) {
		fmt.Printf("Client with ID %s has connected\n", c.Id)
	})

	server.On("login", func(c *gosocketio.Channel, args AuthMessage) {
		if args.Username == "" || args.Password == "" || args.Password != "HelloWorld!!@#" { //Testing password, please change
			c.Emit("FailedLogin", Message{TaskID: 0, DataType: stringType, Data: []byte("Failed")})
		} else {
			// Send to the login function and add the user to the DB
			login(c.Id(), args.Username, args.Password)
			c.Emit("SuccessfulLogin", Message{TaskID: 0, DataType: stringType, Data: []byte("Success")})
		}

	})
}

func login(clientID string, username string, password string) {
	// Authenticate users upon connection
}

func setupDatabase() {
	_, err := os.OpenFile("raven.db", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("Error opening/creating database: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", "raven.db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = db.Exec("CREATE TABLE `users` (`db_id` INTEGER PRIMARY KEY AUTOINCREMENT, `client_id` VARCHAR(64) NULL, `username` VARCHAR(128) NULL, `password` VARCHAR(128) NULL)")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	_, err = db.Exec("CREATE TABLE `clients` ()") //TODO: Add clients table
}
