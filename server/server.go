package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Lazzzer/labo1-sdr/utils"
)

var invalidNbArgsMessage = "Error: Invalid number of arguments. Type 'help' for more information.\n"

// Channels
var eChan = make(chan []utils.Event)
var jChan = make(chan []utils.Job)
var uChan = make(chan []utils.User)

// TODO: Présentation clean
func showEvents() string {
	events := <-eChan
	response := "Events:\n"

	go func() {
		eChan <- events
	}()

	for _, event := range events {
		response += event.Name + "\n"
	}

	return response
}

func showHelp() string {
	var help = "---------------------------------------------------------\n"
	help += "# Description of the command:\nHow to use the command\n \n"
	help += "# Display all commands:\nhelp\n \n"
	help += "# Create an event (will ask your password):\ncreate <eventName> <username> <job1> <nbVolunteer1> [<job2> <nbVolunteer2> ...]\n \n"
	help += "# Close an event (will ask your password):\nclose <idEvent> <username>\n \n"
	help += "# Register to an event (will ask your password):\nregister <idEvent> <idJob> <username>\n \n"
	help += "# Show all events:\nshowAll\n \n"
	help += "# Show all jobs from an event:\nshowJobs <idEvent>\n \n"
	help += "# Show all volunteers from an event:\njobRepartition <idEvent>\n \n"
	help += "# Quit the program:\nquit\n"
	help += "---------------------------------------------------------\n"
	return help
}

func createEvent() string {
	return "create"
}

func verifyUser(username, password string) (utils.User, bool) {
	users := <-uChan

	var returnedUser utils.User
	ok := false

	for _, user := range users {
		if user.Username == username && user.Password == password {
			returnedUser = user
			ok = true
			break
		}
	}

	go func() {
		uChan <- users
	}()

	return returnedUser, ok
}

func register(args []string) string {

	if len(args) != utils.REGISTER.MinArgs {
		return invalidNbArgsMessage
	}

	// idEvent := args[0]
	// idJob := args[1]
	username := args[2]
	password := args[3]

	user, ok := verifyUser(username, password)

	if !ok {
		return "Error: Access denied.\n"
	}

	// TODO : Ajouter l'utilisateur à l'événement
	return "You're in " + user.Username + "! \n"
}

func processCommand(command string) (string, bool) {
	args := strings.Fields(command)

	if len(args) == 0 {
		return "Empty command", false
	}

	var response string
	name := args[0]
	args = args[1:]
	end := false

	switch name {
	case utils.HELP.Name:
		response = showHelp()
	case utils.CREATE.Name:
		response = createEvent()
	case utils.CLOSE.Name:
		response = "TODO"
	case utils.REGISTER.Name:
		response = register(args)
	case utils.SHOW_ALL.Name:
		response = showEvents()
	case utils.SHOW_JOBS.Name:
		response = "TODO"
	case utils.JOBS_REPARTITION.Name:
		response = "TODO"
	case utils.QUIT.Name:
		end = true
	default:
		response = "Error: Invalid command. Type 'help' for a list of commands.\n"
	}

	return response, end
}

func handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		input, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println(err)
			break
		}

		response, end := processCommand(strings.TrimSpace(string(input)))
		fmt.Print(conn.RemoteAddr().String()+" at "+time.Now().Format("15:04:05")+" -> ", string(input))

		if end {
			fmt.Println(conn.RemoteAddr().String() + " disconnected at " + time.Now().Format("15:04:05"))
			break
		}

		conn.Write([]byte(response))
	}
	conn.Close()
}

func main() {
	config := utils.GetConfig("config.json")
	users, events, jobs := utils.GetEntities("entities.json")

	fmt.Println(users)
	fmt.Println(events)
	fmt.Println(jobs)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(config.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go func() {
		uChan <- users
	}()

	go func() {
		eChan <- events
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(conn.RemoteAddr().String() + " connected at " + time.Now().Format("15:04:05"))
		}
		go handleConn(conn)
	}
}
