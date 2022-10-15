package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Lazzzer/labo1-sdr/utils"
	"github.com/Lazzzer/labo1-sdr/utils/types"
)

type Server struct {
	config utils.Config
	eChan  chan []utils.Event
	jChan  chan []utils.Job
	uChan  chan []utils.User
}

// Les méthodes avec des types génériques n'existant pas en Go, on utilise des fonctions qui ne sont pas liées au type Server
func getEntitiesFromChannel[T utils.Event | utils.Job | utils.User](ch <-chan []T, s *Server) []T {
	entities := <-ch
	s.debug(reflect.TypeOf(entities).String(), true, true)

	return entities
}

func loadEntitiesToChannel[T utils.Event | utils.Job | utils.User](ch chan<- []T, entities []T, s *Server) {
	ch <- entities
	s.debug(reflect.TypeOf(entities).String(), true, false)
}

// Debug
func (s *Server) printDebug(title string) {
	if !s.config.Debug {
		fmt.Println(title)

		users := <-s.uChan
		events := <-s.eChan
		jobs := <-s.jChan

		fmt.Print("\nUsers: ")
		fmt.Println(users)
		fmt.Print("\nEvents: ")
		fmt.Println(events)
		fmt.Print("\nJobs: ")
		fmt.Println(jobs)
		fmt.Println()

		s.uChan <- users
		s.eChan <- events
		s.jChan <- jobs
	}
}

func (s *Server) debug(entity string, debug, start bool) {
	if s.config.Debug {
		if start {
			log.Println("DEBUG: using     shared entity: " + entity)
			time.Sleep(4 * time.Second)
		} else {
			log.Println("DEBUG: releasing shared entity: " + entity)
		}
	}
}

// Helpers
func (s *Server) verifyUser(username, password string) (utils.User, bool) {
	users := getEntitiesFromChannel(s.uChan, s)

	var returnedUser utils.User
	ok := false

	for _, user := range users {
		if user.Username == username && user.Password == password {
			returnedUser = user
			ok = true
			break
		}
	}

	loadEntitiesToChannel(s.uChan, users, s)

	return returnedUser, ok
}

func (s *Server) getEventById(id int) (utils.Event, bool) {
	events := getEntitiesFromChannel(s.eChan, s)

	var returnedEvent utils.Event
	ok := false

	for _, event := range events {
		if event.Id == id {
			returnedEvent = event
			ok = true
			break
		}
	}

	loadEntitiesToChannel(s.eChan, events, s)

	return returnedEvent, ok
}

func (s *Server) removeUserInJob(idUser int, job *utils.Job) {
	for i, volunteerId := range job.VolunteerIds {
		if volunteerId == idUser {
			job.VolunteerIds[i] = job.VolunteerIds[len(job.VolunteerIds)-1]
			job.VolunteerIds = job.VolunteerIds[:len(job.VolunteerIds)-1]
			break
		}
	}
}

func (s *Server) addUserToJob(event *utils.Event, idJob, idUser int) (string, bool) {
	jobs := getEntitiesFromChannel(s.jChan, s)

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
				s.removeUserInJob(idUser, &jobs[i])
			}
		}
		jobs[index].VolunteerIds = append(jobs[index].VolunteerIds, idUser)
	}

	loadEntitiesToChannel(s.jChan, jobs, s)

	return errMsg, ok
}

func (s *Server) closeEvent(idEvent, idUser int) (string, bool) {
	events := getEntitiesFromChannel(s.eChan, s)

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

	loadEntitiesToChannel(s.eChan, events, s)

	return errMsg, ok
}

func (s *Server) checkNbArgs(args []string, command *types.Command) (string, bool) {
	if len(args) != command.MinArgs {
		return "Error: Invalid number of arguments. Type 'help' for more information.\n", false
	}
	return "", true
}

// Functions of each command
func (s *Server) showHelp(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.HELP); !ok {
		return msg
	}

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

func (s *Server) createEvent(command []string) string {
	if len(command) < 5 || len(command)%2 != 1 {
		return "TODO"
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

	user, okUser := s.verifyUser(username, password)

	if !okUser {
		return "Error: Access denied."
	}

	jobs := <-s.jChan
	events := <-s.eChan
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

	s.jChan <- jobs

	newEvent := utils.Event{Id: eventId, Name: command[0], CreatorId: user.Id, JobIds: allJobsId}
	events = append(events, newEvent)

	s.eChan <- events

	return "Event created"
}

func (s *Server) close(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.CLOSE); !ok {
		return msg
	}

	idEvent, errEvent := strconv.Atoi(args[0])
	username := args[1]
	password := args[2]

	if errEvent != nil {
		return "Error: event id must be integer.\n"
	}

	user, okUser := s.verifyUser(username, password)
	if !okUser {
		return "Error: Access denied.\n"
	}

	errMsg, ok := s.closeEvent(idEvent, user.Id)

	if !ok {
		return errMsg
	}

	return "Event with id " + strconv.Itoa(idEvent) + " is closed.\n"
}

func (s *Server) register(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.REGISTER); !ok {
		return msg
	}

	idEvent, errEvent := strconv.Atoi(args[0])
	idJob, errJob := strconv.Atoi(args[1])
	username := args[2]
	password := args[3]

	if errEvent != nil || errJob != nil {
		return "Error: Ids must be integers.\n"
	}

	user, okUser := s.verifyUser(username, password)
	if !okUser {
		return "Error: Access denied.\n"
	}

	event, okEvent := s.getEventById(idEvent)
	if !okEvent {
		return "Error: Event not found by this id.\n"
	} else if event.Closed {
		return "Error: Event is closed.\n"
	} else {
		if event.CreatorId == user.Id {
			return "Error: Creator of the event cannot register for a job.\n"
		}
	}

	msg, okJob := s.addUserToJob(&event, idJob, user.Id)

	if !okJob {
		return msg
	}
	return "User " + user.Username + " registered to job with id " + strconv.Itoa(idJob) + " in event " + event.Name + ".\n"
}

// TODO: Présentation clean
func (s *Server) showEvents(args []string) string {
	if msg, ok := s.checkNbArgs(args, &utils.SHOW_ALL); !ok {
		return msg
	}

	events := getEntitiesFromChannel(s.eChan, s)

	response := "Events:\n"

	loadEntitiesToChannel(s.eChan, events, s)

	for _, event := range events {
		response += event.Name + "\n"
	}

	return response
}

// Command processing
func (s *Server) processCommand(command string) (string, bool) {
	args := strings.Fields(command)

	if len(args) == 0 {
		return "Empty command", false
	}

	var response string

	name := args[0]
	args = args[1:]
	end := false

	s.printDebug("\n---------- START COMMAND ----------")

	switch name {
	case utils.HELP.Name:
		response = s.showHelp(args)
	case utils.CREATE.Name:
		response = s.createEvent(args)
	case utils.CLOSE.Name:
		response = s.close(args)
	case utils.REGISTER.Name:
		response = s.register(args)
	case utils.SHOW_ALL.Name:
		response = s.showEvents(args)
	case utils.SHOW_JOBS.Name:
		response = "TODO"
	case utils.JOBS_REPARTITION.Name:
		response = "TODO"
	case utils.QUIT.Name:
		end = true
	default:
		response = "Error: Invalid command. Type 'help' for a list of commands.\n"
	}

	s.printDebug("\n---------- END COMMAND ----------")

	return response, end
}

func (s *Server) handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		input, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println(err)
			break
		}

		response, end := s.processCommand(strings.TrimSpace(string(input)))
		fmt.Print(conn.RemoteAddr().String()+" at "+time.Now().Format("15:04:05")+" -> ", string(input))

		if end {
			fmt.Println(conn.RemoteAddr().String() + " disconnected at " + time.Now().Format("15:04:05"))
			break
		}

		conn.Write([]byte(response))
	}
	conn.Close()
}

func (s *Server) Run() {
	users, events, jobs := utils.GetEntities("entities.json")

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(s.config.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	s.eChan = make(chan []utils.Event, 1)
	s.jChan = make(chan []utils.Job, 1)
	s.uChan = make(chan []utils.User, 1)

	s.uChan <- users
	s.eChan <- events
	s.jChan <- jobs

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(conn.RemoteAddr().String() + " connected at " + time.Now().Format("15:04:05"))
		}
		go s.handleConn(conn)
	}
}

func main() {
	config := utils.GetConfig("config.json")
	server := Server{config: config}
	server.Run()
}
