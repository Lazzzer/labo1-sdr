package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/Lazzzer/labo1-sdr/utils"
)

var ErrorEmptyInput = errors.New("empty input")
var ErrorInvalidInput = errors.New("invalid input")

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
		return "", ErrorEmptyInput
	}

	processedInput := strings.Join(args, " ")

	for _, command := range utils.COMMANDS {
		if args[0] == command.Name && len(args) >= command.MinArgs+1 {
			if command.Auth {
				processedInput += " " + askPassword()
			}
			return processedInput, nil
		}
	}
	return processedInput, ErrorInvalidInput
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

	go func() {
		io.Copy(os.Stdout, conn) // Lecture des réponses du serveur
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		processedInput, err := processInput(input)

		if err != nil {
			if err == ErrorInvalidInput {
				fmt.Println("Error: Invalid command. Type 'help' for a list of commands.")
			}
			continue
		}

		io.Copy(conn, strings.NewReader(processedInput+"\n")) // Passage de l'input traité au serveur

		if processedInput == utils.QUIT.Name {
			fmt.Println("Goodbye!")
			break
		}
	}
}
