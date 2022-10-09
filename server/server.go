package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Lazzzer/labo1-sdr/utils"
)

// Channels
var mChan = make(chan []utils.Manifestation)

func showAllManifestations() string {
	manifestations := <-mChan
	response := "Manifestations:\n"

	go func() {
		mChan <- manifestations
	}()

	for _, manifestation := range manifestations {
		response += manifestation.Name + "\n"
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

func quit() (string, bool) {
	fmt.Println("Fermeture de la connexion")
	return "Goodbye!", true
}

func processCommand(command string) (string, bool) {
	//var allCommands = []string{"help", "create", "close", "register", "showAll", "showJobs", "jobRepartition", "quit"}
	var response string
	end := false
	//splited := strings.Fields(command)
	fmt.Println("splited")
	switch command {
	case "quit":
		fmt.Println("splited")
		response, end = quit()
	case "help":
		response = showHelp()
	case "createEvent":
		response = createEvent()
	case "showAll":
		response = showAllManifestations()
	default:
		response = "Unknown command"
	}

	return response + "\n", end
}

func handleConn(conn net.Conn) {
	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')

		if err != nil {
			fmt.Println(err)
			break
		}

		response, end := processCommand(strings.TrimSpace(string(netData)))

		fmt.Print(conn.RemoteAddr().String()+" at "+time.Now().Format("15:04:05")+" -> ", string(netData))
		conn.Write([]byte(response + "\n"))

		if end {
			break
		}
	}
	conn.Close()
}

func main() {
	config := utils.GetConfig("config.json")
	_, manifestations := utils.GetEntities("entities.json")

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	go func() {
		mChan <- manifestations
	}()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(connection.RemoteAddr().String() + " connected")
		}
		go handleConn(connection)
	}
}
