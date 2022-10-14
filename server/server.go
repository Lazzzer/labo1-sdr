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

var config utils.Config
var invalidNbArgsMessage = "Error: Invalid number of arguments. Type 'help' for more information.\n"

// Channels
var eChan = make(chan []utils.Event, 1)
var jChan = make(chan []utils.Job, 1)
var uChan = make(chan []utils.User, 1)

// Debug
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

	uChan <- users
	eChan <- events
	jChan <- jobs
}

func debug(entity string, debug, start bool) {
	if config.Debug {
		if start {
			log.Println("DEBUG: using     shared entity: " + entity)
			time.Sleep(4 * time.Second)
		} else {
			log.Println("DEBUG: releasing shared entity: " + entity)
		}
	}
}

func verifyUser(username, password string) (utils.User, bool) {
	users := <-uChan
	debug("users", true, true)

	var returnedUser utils.User
	ok := false

	for _, user := range users {
		if user.Username == username && user.Password == password {
			returnedUser = user
			ok = true
			break
		}
	}

	uChan <- users
	debug("users", true, false)

	return returnedUser, ok
}

func getEventById(id int) (utils.Event, bool) {
	events := <-eChan
	debug("events", true, true)

	var returnedEvent utils.Event
	ok := false

	for _, event := range events {
		if event.Id == id {
			returnedEvent = event
			ok = true
			break
		}
	}

	eChan <- events
	debug("events", true, false)

	return returnedEvent, ok
}

func removeUserInJob(idUser int, job *utils.Job) {
	for i, volunteerId := range job.VolunteerIds {
		if volunteerId == idUser {
			job.VolunteerIds[i] = job.VolunteerIds[len(job.VolunteerIds)-1]
			job.VolunteerIds = job.VolunteerIds[:len(job.VolunteerIds)-1]
			break
		}
	}
}

func addUserToJob(event *utils.Event, idJob, idUser int) (string, bool) {
	jobs := <-jChan
	debug("jobs", true, true)

	var index int
	ok := false
	errMsg := ""

	for i, job := range jobs {
		if job.Id == idJob {
			index = i
			break
		}
	}

	if jobs[index].EventId != event.Id {
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
		// Suppression de l'utilisateur dans un job de la manifestation
		for i, job := range jobs {
			if job.EventId == event.Id {
				removeUserInJob(idUser, &jobs[i])
			}
		}
		jobs[index].VolunteerIds = append(jobs[index].VolunteerIds, idUser)
	}

	jChan <- jobs
	debug("jobs", true, false)

	return errMsg, ok
}

func closeEvent(idEvent, idUser int) (string, bool) {
	events := <-eChan
	debug("events", true, true)
	jobs := <-jChan
	debug("jobs", true, true)

	var index int
	ok := false
	found := false
	errMsg := ""

	for i, event := range events {
		if event.Id == idEvent {
			index = i
			found = true
			break
		}
	}

	if !found {
		errMsg = "Error: Event not found with given id.\n"
	} else if events[index].CreatorId != idUser {
		errMsg = "Error: Only the creator of the event can close it.\n"
	} else if events[index].Closed {
		errMsg = "Error: Event is already closed.\n"
	} else {
		events[index].Closed = true
		ok = true
	}

	eChan <- events
	debug("events", true, false)
	jChan <- jobs
	debug("jobs", true, false)

	return errMsg, ok
}

// Command processing

// TODO: Présentation clean
func showEvents() string {
	events := <-eChan
	debug("events", true, true)

	response := "Events:\n"

	eChan <- events
	debug("events", true, false)

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

	idEvent, errEvent := strconv.Atoi(args[0])
	username := args[1]
	password := args[2]

	if errEvent != nil {
		return "Error: event id must be integer.\n"
	}

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

func createEvent(command []string) string {
	if len(command) < 5 || len(command)%2 != 1 {
		return invalidNbArgsMessage
	}

	var nbVolunteersPerJob []int
	var jobsName []string
	for i := 1; i < len(command)-2; i++ {
		if i%2 == 0 {
			if nbVolunteer, err := strconv.Atoi(command[i]); err != nil || nbVolunteer < 0 {
				return "The number of volunteers must be an positive integer."
			} else {
				nbVolunteersPerJob = append(nbVolunteersPerJob, nbVolunteer)
			}
		} else {
			jobsName = append(jobsName, command[i])
		}
	}

	username := command[len(command)-2]
	password := command[len(command)-1]

	user, okUser := verifyUser(username, password)

	if !okUser {
		return "Error: Access denied."
	}

	jobs := <-jChan
	events := <-eChan
	eventId := len(events) + 1
	currentJobId := len(jobs) + 1
	allJobsId := []int{}
	for i := 0; i < len(jobsName); i++ {
		jobs = append(jobs, utils.Job{
			Id:           currentJobId,
			Name:         jobsName[i],
			CreatorId:    nbVolunteersPerJob[i],
			EventId:      user.Id,
			NbVolunteers: eventId,
			VolunteerIds: []int{}})
		allJobsId = append(allJobsId, currentJobId)
		currentJobId++
	}

	jChan <- jobs

	newEvent := utils.Event{Id: eventId, Name: command[0], CreatorId: user.Id, JobIds: allJobsId}
	events = append(events, newEvent)

	eChan <- events

	return "Event created"
}

func register(args []string) string {
	if len(args) != utils.REGISTER.MinArgs {
		return invalidNbArgsMessage
	}

	idEvent, errEvent := strconv.Atoi(args[0])
	idJob, errJob := strconv.Atoi(args[1])
	username := args[2]
	password := args[3]

	if errEvent != nil || errJob != nil {
		return "Error: Ids must be integers.\n"
	}

	user, okUser := verifyUser(username, password)
	if !okUser {
		return "Error: Access denied.\n"
	}

	event, okEvent := getEventById(idEvent)
	if !okEvent {
		return "Error: Event not found by this id.\n"
	} else if event.Closed {
		return "Error: Event is closed.\n"
	} else {
		if event.CreatorId == user.Id {
			return "Error: Creator of the event cannot register for a job.\n"
		}
	}

	msg, okJob := addUserToJob(&event, idJob, user.Id)

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

	// printDebug("\n---------- START COMMAND ----------")

	switch name {
	case utils.HELP.Name:
		response = showHelp()
	case utils.CREATE.Name:
		response = createEvent(args)
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

	// printDebug("\n---------- END COMMAND ----------")

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
	config = utils.GetConfig("config.json")
	users, events, jobs := utils.GetEntities("entities.json")

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(config.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	uChan <- users
	eChan <- events
	jChan <- jobs

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
