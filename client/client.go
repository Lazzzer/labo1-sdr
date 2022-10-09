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

func askPassword() string {
	var password string
	fmt.Print("Password: ")
	// TODO: Lecture de la saisie sans afficher les caractères
	fmt.Scanln(&password)
	return password
}

func processInput(input string) (string, error) {
	args := strings.Fields(input)

	if len(args) == 0 {
		return "", fmt.Errorf("invalid command")
	}
	processedInput := strings.Join(args, " ")

	for _, command := range utils.COMMANDS {
		if args[0] == command.Name && len(args) >= command.MinArgs+1 {
			if command.Auth {
				password := askPassword()
				processedInput += " " + password
			}
			return processedInput, nil
		}
	}
	return processedInput, fmt.Errorf("invalid command")
}

func main() {
	config := utils.GetConfig("config.json")

	conn, err := net.Dial("tcp", config.Host+":"+strconv.Itoa(config.Port))

	if err != nil {
		panic(err)
	} else {
		fmt.Println("Welcome! Please enter a command.\nType 'help' for a list of commands.")
	}

	defer conn.Close()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		input, _ := reader.ReadString('\n')

		newInput, err := processInput(input)

		if err != nil {
			fmt.Println("Error: Invalid command. Type 'help' for a list of commands.")
			continue
		}

		fmt.Fprintf(conn, newInput+"\n")

		var lines []string
		scanner := bufio.NewScanner(conn)
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

		if strings.TrimSpace(string(newInput)) == "quit" {
			fmt.Println("Closing connection")
			break
		}
	}
}
