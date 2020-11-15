package main

import (
	"encoding/gob"
	"fmt"
	"net"
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

func createProcess(uuid int) (new_process *Process) {
	new_process = new(Process)
	new_process.Uuid = uuid
	new_process.Assigned = false
	new_process.Progress = 0
	return
}

type Server struct {
	port            int
	process_asigned int
	processes       []Process
}

func (self *Server) init(process_capacity int, port int) {
	if port == 0 {
		port = 3000 // otra babosada mas de go -_-
	}
	self.process_asigned = 0
	self.port = port

	for h := 0; h < process_capacity; h++ {
		self.processes = append(self.processes, *(createProcess(len(self.processes))))
	}
}

func (self *Server) start() {
	for {
		time.Sleep(time.Second / 2)
		if self.process_asigned >= len(self.processes) {
			continue
		}

		for h := 0; h < len(self.processes); h++ {
			if self.processes[h].Assigned {
				continue
			}

			fmt.Println(self.processes[h].toString())
			self.processes[h].Progress++
		}
		fmt.Println("==========================")
	}
}

func panicIfErr(e error) {
	if e != nil {
		panic(e)
	}
}

func (self *Server) handelConnection(connection net.Conn) {
	var transmiter *gob.Encoder = gob.NewEncoder(connection)
	var reciver *gob.Decoder = gob.NewDecoder(connection)
	var response string

	panicIfErr(reciver.Decode(&response))
	if response == "GET" {
		for h := 0; h < len(self.processes); h++ {
			if !self.processes[h].Assigned {
				self.processes[h].Assigned = true
				self.process_asigned++
				err := gob.NewEncoder(connection).Encode(self.processes[h])
				if err != nil {
					fmt.Println(err)
					panic(fmt.Sprintf("failed to respond to client %s", connection.LocalAddr().String()))
				}
				return
			}
		}
	} else if response == "POST" {
		defer connection.Close()
		var forgotten_process *Process
		panicIfErr(transmiter.Encode("ok"))
		panicIfErr(reciver.Decode(&forgotten_process))

		for h := 0; h < len(self.processes); h++ {
			if forgotten_process.Uuid == self.processes[h].Uuid {
				self.processes[h].Assigned = false
				self.processes[h].Progress = forgotten_process.Progress
				return
			}
		}
	}
}

func (self *Server) lisent() {
	lisenter, err := net.Listen("tcp", fmt.Sprintf(":%d", self.port))

	if err != nil {
		panic(fmt.Sprintf("Fail to connect to port %d\n", self.port))
	}

	fmt.Printf("Listing on port :%d\n", self.port)

	for self.process_asigned < len(self.processes) {
		connection, err := lisenter.Accept()
		if err != nil {
			fmt.Printf("Failed to retrive connection from %s", connection.RemoteAddr().String())
			continue
		}
		go self.handelConnection(connection)
	}
}

func main() {
	var tcp_server *Server = new(Server)
	tcp_server.init(5, 0)
	go tcp_server.lisent()
	go tcp_server.start()

	var trash string
	for {
		fmt.Scanln(&trash)
		if trash == "q" {
			break
		}
	}

}
