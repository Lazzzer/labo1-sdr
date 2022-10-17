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
	fmt.Println(utils.MESSAGE.LoginStart)
	defer fmt.Println(utils.MESSAGE.LoginEnd)

	fmt.Print(utils.BOLD + "Enter Username: " + utils.RESET)
	username, errUsername := bufio.NewReader(os.Stdin).ReadString('\n')
	usernameArr := strings.Fields(username)

	if errUsername != nil || len(usernameArr) != 1 {
		return "", fmt.Errorf("invalid username")
	}
	username = usernameArr[0]

	fmt.Print(utils.BOLD + "Enter Password: " + utils.RESET)
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
		log.Fatal("❌ " + utils.RED + "Could not connect to the server." + utils.RESET)
	} else {
		fmt.Println(utils.MESSAGE.Title)
	}

	defer conn.Close()

	go func() {
		_, errFrom := io.Copy(os.Stdout, conn) // Lecture des réponses du serveur
		if errFrom != nil {
			os.Exit(1)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		processedInput, err := c.processInput(input)

		if err != nil {
			fmt.Print(utils.MESSAGE.Error.InvalidCommand)
			continue
		}

		io.Copy(conn, strings.NewReader(processedInput+"\n")) // Passage de l'input traité au serveur

		if processedInput == utils.QUIT.Name {
			fmt.Println(utils.MESSAGE.Goodbye)
			break
		}
	}
}
