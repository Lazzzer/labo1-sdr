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

func handleConnection(connection net.Conn, uMap *sync.Map, mMap *sync.Map) {
	for {
		netData, err := bufio.NewReader(connection).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}

		var response string

		if strings.TrimSpace(string(netData)) == "quit" {
			fmt.Println("Closing connection to " + connection.LocalAddr().String())
			break
		}
		if strings.TrimSpace(string(netData)) == "showAll" {
			fmt.Println("Showing all manifestations")
			mMap.Range(func(key, value interface{}) bool {
				fmt.Println(value)
				response = response + value.(utils.Manifestation).Name + "\n"
				return true
			})
		}

		fmt.Print("From Client "+connection.LocalAddr().String()+" -> ", string(netData))
		connection.Write([]byte(response + "\n"))
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
