package server

import (
	"bufio"
	_ "embed"
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

//go:embed entities.json
var entities string

type Server struct {
	Config utils.Config
	eChan  chan map[int]utils.Event
	jChan  chan map[int]utils.Job
	uChan  chan map[int]utils.User
}

// Les méthodes avec des types génériques n'existant pas en Go, on utilise des fonctions qui ne sont pas liées au type Server
func getEntitiesFromChannel[T utils.Event | utils.Job | utils.User](ch <-chan map[int]T, s *Server) map[int]T {
	entities := <-ch
	s.debug(reflect.TypeOf(entities).String(), true, true)

	return entities
}

func loadEntitiesToChannel[T utils.Event | utils.Job | utils.User](ch chan<- map[int]T, entities map[int]T, s *Server) {
	ch <- entities
	s.debug(reflect.TypeOf(entities).String(), true, false)
}

// Debug
func (s *Server) printDebug(title string) {
	if !s.Config.Debug && !s.Config.Silent {
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
	if s.Config.Debug {
		if start {
			log.Println("DEBUG: using     shared entity: " + entity)
			time.Sleep(4 * time.Second)
		} else {
			log.Println("DEBUG: releasing shared entity: " + entity)
		}
	}
}

// Helpers
func (s *Server) verifyUser(username, password string) (int, bool) {
	users := getEntitiesFromChannel(s.uChan, s)
	defer loadEntitiesToChannel(s.uChan, users, s)

	for key, user := range users {
		if user.Username == username && user.Password == password {
			return key, true
		}
	}

	return 0, false
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

func (s *Server) addUserToJob(idEvent, idJob, idUser int) (string, bool) {
	jobs := getEntitiesFromChannel(s.jChan, s)
	defer loadEntitiesToChannel(s.jChan, jobs, s)

	job, ok := jobs[idJob]

	if ok {
		// Différentes vérifications selon le cahier des charges avec les messages d'erreur correspondants
		if job.EventId != idEvent {
			return "Error: Given event id does not match id in job.\n", false
		} else if job.CreatorId == idUser {
			return "Error: Creator of the event cannot register for a job.\n", false
		} else if len(job.VolunteerIds) == job.NbVolunteers {
			return "Error: Job is already full.\n", false
		} else {
			for _, id := range job.VolunteerIds {
				if id == idUser {
					return "Error: User is already registered in this job.\n", false
				}
			}
		}

		// Suppression de l'utilisateur dans un job de la manifestation
		for exploredJobId, exploredJob := range jobs {
			if exploredJob.EventId == idEvent {
				s.removeUserInJob(idUser, &exploredJob)
				jobs[exploredJobId] = exploredJob
			}
		}
		// Ajout de l'utilisateur dans son nouveau job
		job.VolunteerIds = append(job.VolunteerIds, idUser)
		jobs[idJob] = job
	} else {
		return "Error: Job not found with given id.\n", false
	}

	return "", true
}

func (s *Server) closeEvent(idEvent, idUser int) (string, bool) {
	events := getEntitiesFromChannel(s.eChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)

	event, okEvent := events[idEvent]

	if !okEvent {
		return "Error: Event not found with given id.\n", false
	} else if event.CreatorId != idUser {
		return "Error: Only the creator of the event can close it.\n", false
	} else if event.Closed {
		return "Error: Event is already closed.\n", false
	} else {
		event.Closed = true
		events[idEvent] = event
	}

	return "", true
}

func (s *Server) checkNbArgs(args []string, command *types.Command, optional bool) (string, bool) {
	msg := "Error: Invalid number of arguments. Type 'help' for more information.\n"
	if optional {
		if len(args) < command.MinArgs || len(args)%command.MinOptArgs != 1 {
			return msg, false
		}
	} else {
		if len(args) != command.MinArgs && len(args)%command.MinOptArgs != 1 {
			return msg, false
		}
	}

	return "", true
}

// Functions of each command
func (s *Server) showHelp(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.HELP, false); !ok {
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

func (s *Server) createEvent(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.CREATE, true); !ok {
		return msg
	}

	username := args[len(args)-2]
	password := args[len(args)-1]

	userId, okUser := s.verifyUser(username, password)

	if !okUser {
		return "Error: Access denied."
	}

	var nbVolunteersPerJob []int
	var jobsName []string

	for i := 1; i < len(args)-utils.CREATE.MinOptArgs; i++ {
		if i%utils.CREATE.MinOptArgs == 0 {
			if nbVolunteer, err := strconv.Atoi(args[i]); err != nil || nbVolunteer < 0 {
				return "Error: The number of volunteers must be an positive integer."
			} else {
				nbVolunteersPerJob = append(nbVolunteersPerJob, nbVolunteer)
			}
		} else {
			jobsName = append(jobsName, args[i])
		}
	}

	jobs := getEntitiesFromChannel(s.jChan, s)
	defer loadEntitiesToChannel(s.jChan, jobs, s)

	events := getEntitiesFromChannel(s.eChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)

	eventId := len(events) + 1
	currentJobId := len(jobs) + 1
	allJobsId := []int{}
	for i := 0; i < len(jobsName); i++ {
		jobs[currentJobId] = utils.Job{
			Name:         jobsName[i],
			CreatorId:    nbVolunteersPerJob[i],
			EventId:      userId,
			NbVolunteers: eventId,
			VolunteerIds: []int{}}
		allJobsId = append(allJobsId, currentJobId)
		currentJobId++
	}

	newEvent := utils.Event{Name: args[0], CreatorId: userId, JobIds: allJobsId}
	events[eventId] = newEvent

	return "Event with id " + strconv.Itoa(eventId) + " and " + strconv.Itoa(len(allJobsId)) + " job(s) " + " created\n"
}

func (s *Server) close(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.CLOSE, false); !ok {
		return msg
	}

	idEvent, errEvent := strconv.Atoi(args[0])
	username := args[1]
	password := args[2]

	if errEvent != nil {
		return "Error: event id must be integer.\n"
	}

	userId, okUser := s.verifyUser(username, password)
	if !okUser {
		return "Error: Access denied.\n"
	}

	errMsg, ok := s.closeEvent(idEvent, userId)

	if !ok {
		return errMsg
	}

	return "Event with id " + strconv.Itoa(idEvent) + " is closed.\n"
}

func (s *Server) register(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.REGISTER, false); !ok {
		return msg
	}

	idEvent, errEvent := strconv.Atoi(args[0])
	idJob, errJob := strconv.Atoi(args[1])
	username := args[2]
	password := args[3]

	if errEvent != nil || errJob != nil {
		return "Error: Ids must be integers.\n"
	}

	userId, okUser := s.verifyUser(username, password)
	if !okUser {
		return "Error: Access denied.\n"
	}

	events := getEntitiesFromChannel(s.eChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)

	event, okEvent := events[idEvent]

	if !okEvent {
		return "Error: Event not found by this id.\n"
	} else if event.Closed {
		return "Error: Event is closed.\n"
	} else {
		if event.CreatorId == userId {
			return "Error: Creator of the event cannot register for a job.\n"
		}
	}

	msg, okJob := s.addUserToJob(idEvent, idJob, userId)

	if !okJob {
		return msg
	}
	return "User registered to job with id " + strconv.Itoa(idJob) + " in event " + event.Name + ".\n"
}

// TODO: Présentation clean
func (s *Server) showEvents(args []string) string {
	if msg, ok := s.checkNbArgs(args, &utils.SHOW_ALL, false); !ok {
		return msg
	}

	events := getEntitiesFromChannel(s.eChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)

	response := "Events:\n"

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
		if !s.Config.Silent {
			fmt.Print(conn.RemoteAddr().String()+" at "+time.Now().Format("15:04:05")+" -> ", string(input))
		}

		if end {
			if !s.Config.Silent {
				fmt.Println(conn.RemoteAddr().String() + " disconnected at " + time.Now().Format("15:04:05"))
			}
			break
		}

		conn.Write([]byte(response))
	}
	conn.Close()
}

func (s *Server) Run() {

	users, events, jobs := utils.GetEntities(entities)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(s.Config.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	s.eChan = make(chan map[int]utils.Event, 1)
	s.jChan = make(chan map[int]utils.Job, 1)
	s.uChan = make(chan map[int]utils.User, 1)

	s.uChan <- users
	s.eChan <- events
	s.jChan <- jobs

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		} else {
			if !s.Config.Silent {
				fmt.Println(conn.RemoteAddr().String() + " connected at " + time.Now().Format("15:04:05"))
			}
		}
		go s.handleConn(conn)
	}
}
