#!/bin/bash

if (( $EUID != 0 )); then
    echo "Please run this with sudo or as root"
    exit
fi

apt-get install golang-go
go get -u github.com/kabukky/httpscerts
go get -u github.com/gorilla/websocket
go get -u github.com/zhuangsirui/binpacker

go build -o server main.go handler.go userhandler.go
echo "Run ./server -help for options"
echo "Done"