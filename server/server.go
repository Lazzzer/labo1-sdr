package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Lazzzer/labo1-sdr/utils"
)

func main() {
	config := utils.GetConfig("config.json")
	fmt.Println(config)
	users, manifestations := utils.GetEntities("entities.json")
	fmt.Println(users, manifestations)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	connection, err := listener.Accept()
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(connection.LocalAddr().String() + " connected")
	}

	for {
		netData, err := bufio.NewReader(connection).ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		if strings.TrimSpace(string(netData)) == "quit" {
			fmt.Println("Closing connection to " + connection.LocalAddr().String())
			return
		}

		fmt.Print("From Client "+connection.LocalAddr().String()+" -> ", string(netData))
		connection.Write([]byte(netData + "\n"))
	}

}
