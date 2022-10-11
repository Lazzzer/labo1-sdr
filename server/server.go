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

func printDebug(title string) {
	fmt.Println(title)

	users := <-uChan
	events := <-eChan
	jobs := <-jChan

	fmt.Print("\nUsers: ")
	fmt.Println(users)
	fmt.Print("\nEvents: ")
	fmt.Println(events)
	fmt.Print("\nJobs: ")
	fmt.Println(jobs)
	fmt.Println()

	go func() {
		uChan <- users
	}()

	go func() {
		eChan <- events
	}()

	go func() {
		jChan <- jobs
	}()
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

func getEventById(id int) (utils.Event, bool) {
	events := <-eChan

	var returnedEvent utils.Event
	ok := false

	for _, event := range events {
		if event.Id == id {
			returnedEvent = event
			ok = true
			break
		}
	}

	go func() {
		eChan <- events
	}()

	return returnedEvent, ok
}

func addUserToJob(idJob, idEvent, idUser int) (string, bool) {
	jobs := <-jChan

	var index int
	ok := false
	errMsg := ""

	for i, job := range jobs {
		if job.Id == idJob {
			index = i
			break
		}
	}

	if jobs[index].EventId != idEvent {
		errMsg = "Error: Given event id does not match id in job.\n"
	} else if jobs[index].CreatorId == idUser {
		errMsg = "Error: Creator of the event cannot register for a job.\n"
	} else if len(jobs[index].VolunteerIds) == jobs[index].NbVolunteers {
		errMsg = "Error: Job is already full.\n"
	} else {
		ok = true
		for _, id := range jobs[index].VolunteerIds {
			if id == idUser {
				ok = false
				errMsg = "Error: User is already registered in this job.\n"
				break
			}
		}
	}

	if ok {
		jobs[index].VolunteerIds = append(jobs[index].VolunteerIds, idUser)
	}

	go func() {
		jChan <- jobs
	}()

	return errMsg, ok
}

func closeEvent(idEvent, idUser int) (string, bool) {
	events := <-eChan
	jobs := <-jChan

	var newJobs []utils.Job
	var newEvents []utils.Event
	var index int
	ok := false
	errMsg := ""

	for i, event := range events {
		if event.Id == idEvent {
			index = i
			ok = true
		} else {
			newEvents = append(newEvents, event)
		}
	}

	if !ok {
		errMsg = "Error: Event not found with given id.\n"
	} else if events[index].CreatorId != idUser {
		errMsg = "Error: Only the creator of the event can delete it.\n"
		ok = false
	} else {
		for _, job := range jobs {
			if job.EventId != idEvent {
				newJobs = append(newJobs, job)
			}
		}
	}

	if !ok {
		go func() {
			eChan <- events
		}()

		go func() {
			jChan <- jobs
		}()
	} else {
		go func() {
			eChan <- newEvents
		}()

		go func() {
			jChan <- newJobs
		}()
	}

	return errMsg, ok

}

// Command processing

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

func close(args []string) string {
	if len(args) != utils.CLOSE.MinArgs {
		return invalidNbArgsMessage
	}

	idEvent, _ := strconv.Atoi(args[0])
	username := args[1]
	password := args[2]

	user, okUser := verifyUser(username, password)
	if !okUser {
		return "Error: Access denied.\n"
	}

	errMsg, ok := closeEvent(idEvent, user.Id)

	if !ok {
		return errMsg
	}

	return "Event with id " + strconv.Itoa(idEvent) + " is closed.\n"
}

func createEvent() string {
	return "create"
}

func register(args []string) string {
	if len(args) != utils.REGISTER.MinArgs {
		return invalidNbArgsMessage
	}

	idEvent, _ := strconv.Atoi(args[0])
	idJob, _ := strconv.Atoi(args[1])
	username := args[2]
	password := args[3]

	user, okUser := verifyUser(username, password)
	if !okUser {
		return "Error: Access denied.\n"
	}

	event, okEvent := getEventById(idEvent)
	if !okEvent {
		return "Error: Event not found by this id.\n"
	} else {
		if event.CreatorId == user.Id {
			return "Error: Creator of the event cannot register for a job.\n"
		}
	}

	msg, okJob := addUserToJob(idJob, idEvent, user.Id)

	if !okJob {
		return msg
	}
	return "User " + user.Username + " registered to job with id " + strconv.Itoa(idJob) + " in event " + event.Name + ".\n"
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

	printDebug("\n---------- START COMMAND ----------")

	switch name {
	case utils.HELP.Name:
		response = showHelp()
	case utils.CREATE.Name:
		response = "TODO"
	case utils.CLOSE.Name:
		response = close(args)
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

	printDebug("\n---------- END COMMAND ----------")

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

	go func() {
		jChan <- jobs
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
