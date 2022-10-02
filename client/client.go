package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/Lazzzer/labo1-sdr/utils"
)

func main() {
	config := utils.GetConfig("config.json")
	fmt.Println(config)

	connection, err := net.Dial("tcp", config.Host+":"+strconv.Itoa(config.Port))

	if err != nil {
		panic(err)
	} else {
		fmt.Println("Welcome! Please enter a command.\nType 'help' for a list of commands.")
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(connection, text+"\n")

		// TODO: Use a Scanner (?) instead of ReadString to read multiple lines
		message, _ := bufio.NewReader(connection).ReadString('\n')
		fmt.Print("From Server -> " + message)

		if strings.TrimSpace(string(text)) == "quit" {
			fmt.Println("Closing connection")
			return
		}
	}
}
