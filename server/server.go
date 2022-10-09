package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Lazzzer/labo1-sdr/utils"
)

// Channels
var eChan = make(chan []utils.Event)

func showEvents() string {
	events := <-eChan
	response := "Events:\n"

	go func() {
		eChan <- events
	}()

	for _, event := range events {
		response += event.Name + "\n"
	}

	return response
}

func showHelp() string {
	var help = "---------------------------------------------------------\n"
	help += "# Description of the command:\nHow to use the command\n \n"
	help += "# Display all commands:\nhelp\n \n"
	help += "# Create an event (will ask your password):\ncreate <eventName> <username> <job1> <nbVolunteer1> [<job2> <nbVolunteer2> ...]\n \n"
	help += "# Close an event (will ask your password):\nclose <idEvent> <username>\n \n"
	help += "# Register to an event (will ask your password):\nregister <idEvent> <idJob> <username>\n \n"
	help += "# Show all events:\nshowAll\n \n"
	help += "# Show all jobs from an event:\nshowJobs <idEvent>\n \n"
	help += "# Show all volunteers from an event:\njobRepartition <idEvent>\n \n"
	help += "# Quit the program:\nquit\n"
	help += "---------------------------------------------------------\n"
	return help
}

func createEvent() string {
	return "create"
}

func processCommand(command string) (string, bool) {
	//var allCommands = []string{"help", "create", "close", "register", "showAll", "showJobs", "jobRepartition", "quit"}
	var response string
	end := false
	//splited := strings.Fields(command)
	switch command {
	case "quit":
		end = true
	case "help":
		response = showHelp()
	case "createEvent":
		response = createEvent()
	case "showAll":
		response = showEvents()
	default:
		response = "Error: Invalid command. Type 'help' for a list of commands.\n"
	}

	return response, end
}

func handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		input, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println(err)
			break
		}

		response, end := processCommand(strings.TrimSpace(string(input)))
		fmt.Print(conn.RemoteAddr().String()+" at "+time.Now().Format("15:04:05")+" -> ", string(input))

		if end {
			fmt.Println(conn.RemoteAddr().String() + " disconnected at " + time.Now().Format("15:04:05"))
			break
		}

		conn.Write([]byte(response))
	}
	conn.Close()
}

func main() {
	config := utils.GetConfig("config.json")
	_, events := utils.GetEntities("entities.json")

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(config.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go func() {
		eChan <- events
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(conn.RemoteAddr().String() + " connected at " + time.Now().Format("15:04:05"))
		}
		go handleConn(conn)
	}
}
