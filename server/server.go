package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Lazzzer/labo1-sdr/utils"
)

func showAllManifestations(mMap *sync.Map) string {
	response := "Manifestations:\n"
	mMap.Range(func(key, value interface{}) bool {
		// fmt.Println(value)
		response = response + value.(utils.Manifestation).Name + "\n"
		return true
	})

	return response
}

// TODO : Add help
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

func quit() (string, bool) {
	fmt.Println("Fermeture de la connexion")
	return "Au revoir!", true
}

func processCommand(command string, m *sync.Map) (string, bool) {
	var response string
	end := false
	switch command {
	case "quit":
		response, end = quit()
	case "help":
		response = showHelp()
	case "showAll":
		response = showAllManifestations(m)
	default:
		response = "Unknown command"
	}

	return response + "\n", end
}

func handleConnection(connection net.Conn, uMap *sync.Map, mMap *sync.Map) {
	for {
		netData, err := bufio.NewReader(connection).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}

		response, end := processCommand(strings.TrimSpace(string(netData)), mMap)

		fmt.Print(connection.RemoteAddr().String()+" at "+time.Now().Format("15:04:05")+" -> ", string(netData))
		connection.Write([]byte(response + "\n"))

		if end {
			break
		}
	}
	connection.Close()
}

func main() {
	config := utils.GetConfig("config.json")
	users, manifestations := utils.GetEntities("entities.json")

	userMap := sync.Map{}
	manifMap := sync.Map{}

	for _, user := range users {
		userMap.Store(user.Username, user) // TODO: Add id for a better key?
	}

	for _, manifestation := range manifestations {
		manifMap.Store(manifestation.Id, manifestation)
	}

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(connection.RemoteAddr().String() + " connected")
		}
		go handleConnection(connection, &userMap, &manifMap)
	}
}
