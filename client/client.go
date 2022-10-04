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

		var lines []string
		scanner := bufio.NewScanner(connection)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) == 0 {
				break
			}
			lines = append(lines, line)
		}
		for _, line := range lines {
			fmt.Println(line)
		}

		if strings.TrimSpace(string(text)) == "quit" {
			fmt.Println("Closing connection")
			return
		}
	}
}
