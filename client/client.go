package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/Lazzzer/labo1-sdr/utils"
	"golang.org/x/term"
)

type Client struct {
	config utils.Config
}

func (c *Client) askCredentials() (string, error) {
	fmt.Println("Enter Username: ")
	username, errUsername := bufio.NewReader(os.Stdin).ReadString('\n')
	usernameArr := strings.Fields(username)

	if errUsername != nil || len(usernameArr) != 1 {
		return "", fmt.Errorf("invalid username")
	}
	username = usernameArr[0]

	fmt.Println("Enter Password: ")
	bytePassword, errPassword := term.ReadPassword(int(syscall.Stdin))

	if errPassword != nil {
		return "", errPassword
	}

	return username + " " + string(bytePassword), nil
}

func (c *Client) processInput(input string) (string, error) {
	args := strings.Fields(input)

	if len(args) == 0 {
		return "", fmt.Errorf("empty input")
	}

	processedInput := strings.Join(args, " ")

	for _, command := range utils.COMMANDS {
		if args[0] == command.Name {
			if command.Auth {
				credentials, err := c.askCredentials()
				if err != nil {
					return "", fmt.Errorf("invalid input")
				}
				processedInput += " " + credentials
			}
			return processedInput, nil
		}
	}
	return "", fmt.Errorf("invalid input")
}

func (c *Client) Run() {
	conn, err := net.Dial("tcp", c.config.Host+":"+strconv.Itoa(c.config.Port))

	if err != nil {
		log.Fatal(err)
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
		processedInput, err := c.processInput(input)

		if err != nil {
			fmt.Println("Error: Invalid input. Type 'help' for a list of commands.")
			continue
		}

		io.Copy(conn, strings.NewReader(processedInput+"\n")) // Passage de l'input traité au serveur

		if processedInput == utils.QUIT.Name {
			fmt.Println("Goodbye!")
			break
		}
	}
}

func main() {
	config := utils.GetConfig("config.json")
	client := Client{config}
	client.Run()
}
