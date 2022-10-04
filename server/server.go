package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/Lazzzer/labo1-sdr/utils"
)

func showAllManifestations(mMap *sync.Map) string {
	fmt.Println("Showing all manifestations")
	response := "Manifestations:\n"
	mMap.Range(func(key, value interface{}) bool {
		fmt.Println(value)
		response = response + value.(utils.Manifestation).Name + "\n"
		return true
	})

	return response
}

// TODO : Add help
func showHelp() string {
	return "help"
}

func processCommand(command string, m *sync.Map) (string, bool) {
	var response string
	end := false
	switch command {
	case "quit":
		fmt.Println("Closing connection")
		response = "Good bye!"
		end = true
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

		fmt.Print("From Client "+connection.LocalAddr().String()+" -> ", string(netData))
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
			fmt.Println(connection.LocalAddr().String() + " connected")
		}
		go handleConnection(connection, &userMap, &manifMap)
	}
}
