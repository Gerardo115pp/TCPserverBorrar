package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Process struct {
	Uuid     int
	Progress int
	Assigned bool
}

func (self Process) toString() string {
	return fmt.Sprintf("%d : %d", self.Uuid, self.Progress)
}

func requestProcess() (assigned_process *Process) {
	connection, err := net.Dial("tcp", ":3000")
	defer connection.Close()
	if err != nil {
		panic(fmt.Sprintf("failed to stablish a connection with the server %s", connection.RemoteAddr().String()))
	}

	err = gob.NewEncoder(connection).Encode("GET")
	if err != nil {
		panic(fmt.Sprintf("failed to send a GET request to the server %s", connection.RemoteAddr().String()))
	}
	err = gob.NewDecoder(connection).Decode(&assigned_process)
	if err != nil {
		panic(err)
	}
	return
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func disownProcess(process *Process) {
	connection, err := net.Dial("tcp", ":3000")
	panicIfErr(err)

	var trasmiter *gob.Encoder = gob.NewEncoder(connection)
	var receiver *gob.Decoder = gob.NewDecoder(connection)
	var response string

	panicIfErr(trasmiter.Encode("POST"))
	panicIfErr(receiver.Decode(&response))
	if response == "ok" {
		panicIfErr(trasmiter.Encode(process))
	}
	return
}

func startProcessing() {
	var assigned_process *Process = requestProcess()
	var signal_lisenter chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(signal_lisenter, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-signal_lisenter:
			fmt.Println("Returning Process")
			disownProcess(assigned_process)
			return
		default:
			fmt.Println(assigned_process.toString())

			assigned_process.Progress++
		}
		time.Sleep(time.Second / 2)
	}

}

func main() {
	startProcessing()
}
