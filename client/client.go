package client

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
	Config utils.Config
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
	conn, err := net.Dial("tcp", c.Config.Host+":"+strconv.Itoa(c.Config.Port))

	if err != nil {
		log.Fatal(err)
	} else {

		title := "  _____                 _     __  __                                   \n"
		title += " | ____|_   _____ _ __ | |_  |  \\/  | __ _ _ __   __ _  __ _  ___ _ __ \n"
		title += " |  _| \\ \\ / / _ \\ '_ \\| __| | |\\/| |/ _` | '_ \\ / _` |/ _` |/ _ \\ '__|\n"
		title += " | |___ \\ V /  __/ | | | |_  | |  | | (_| | | | | (_| | (_| |  __/ |   \n"
		title += " |_____| \\_/ \\___|_| |_|\\__| |_|  |_|\\__,_|_| |_|\\__,_|\\__, |\\___|_|   \n"
		title += "                                                       |___/           "

		fmt.Println(title)
		fmt.Println("Labo 1 SDR - Jonathan Friedli & Lazar Pavicevic")
		fmt.Println("\nWelcome! Please enter a command.\nType 'help' for a list of commands.")
		fmt.Println()
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
