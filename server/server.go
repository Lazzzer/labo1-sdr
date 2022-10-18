package server

import (
	"bufio"
	_ "embed"
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
	Config types.Config
	eChan  chan map[int]types.Event
	uChan  chan map[int]types.User
}

// Les méthodes avec des types génériques n'existant pas en Go, on utilise des fonctions qui ne sont pas liées au type Server
func getEntitiesFromChannel[T types.Event | types.User](ch <-chan map[int]T, s *Server) map[int]T {
	entities := <-ch
	s.debug(reflect.TypeOf(&entities).Elem().Elem().String(), true, true)

	return entities
}

func loadEntitiesToChannel[T types.Event | types.User](ch chan<- map[int]T, entities map[int]T, s *Server) {
	ch <- entities
	s.debug(reflect.TypeOf(&entities).Elem().Elem().String(), true, false)
}

// Debug
// func (s *Server) printDebug(title string) {
// 	if !s.Config.Debug && !s.Config.Silent {
// 		fmt.Println(title)

// 		users := <-s.uChan
// 		events := <-s.eChan

// 		fmt.Print("\nUsers: ")
// 		fmt.Println(users)
// 		fmt.Print("\nEvents: ")
// 		fmt.Println(events)
// 		fmt.Println()

// 		s.uChan <- users
// 		s.eChan <- events
// 	}
// }

func (s *Server) debug(entity string, debug, start bool) {
	if s.Config.Debug {
		if start {
			log.Println(utils.ORANGE + "(DEBUG) " + utils.RED + "ACCESSING" + utils.ORANGE + " shared section for entity: " + entity + utils.RESET)
			time.Sleep(4 * time.Second)
		} else {
			log.Println(utils.ORANGE + "(DEBUG) " + utils.GREEN + "RELEASING" + utils.ORANGE + " shared section for entity: " + entity + utils.RESET)
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

func (s *Server) removeUserInJob(idUser int, job *types.Job) bool {
	for i, volunteerId := range job.VolunteerIds {
		if volunteerId == idUser {
			job.VolunteerIds[i] = job.VolunteerIds[len(job.VolunteerIds)-1]
			job.VolunteerIds = job.VolunteerIds[:len(job.VolunteerIds)-1]
			return true
		}
	}
	return false
}

func (s *Server) addUserToJob(event *types.Event, idJob, idUser int) (string, bool) {

	job, ok := event.Jobs[idJob]

	if ok {
		// Différentes vérifications selon le cahier des charges avec les messages d'erreur correspondants
		if job.CreatorId == idUser {
			return utils.MESSAGE.Error.CreatorRegister, false
		} else if len(job.VolunteerIds) == job.NbVolunteers {
			return utils.MESSAGE.Error.JobFull, false
		} else {
			for _, id := range job.VolunteerIds {
				if id == idUser {
					return utils.MESSAGE.Error.AlreadyRegistered, false
				}
			}
		}

		// Suppression de l'utilisateur dans un job de la manifestation
		for exploredJobId, exploredJob := range event.Jobs {
			if s.removeUserInJob(idUser, &exploredJob) {
				event.Jobs[exploredJobId] = exploredJob
			}
		}
		// Ajout de l'utilisateur dans son nouveau job
		job.VolunteerIds = append(job.VolunteerIds, idUser)
		event.Jobs[idJob] = job
	} else {
		return utils.MESSAGE.Error.JobNotFound, false
	}

	return "", true
}

func (s *Server) closeEvent(idEvent, idUser int) (string, bool) {
	events := getEntitiesFromChannel(s.eChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)

	event, okEvent := events[idEvent]

	if !okEvent {
		return utils.MESSAGE.Error.EventNotFound, false
	} else if event.CreatorId != idUser {
		return utils.MESSAGE.Error.NotCreator, false
	} else if event.Closed {
		return utils.MESSAGE.Error.AlreadyClosed, false
	} else {
		event.Closed = true
		events[idEvent] = event
	}

	return "", true
}

func (s *Server) checkNbArgs(args []string, command *types.Command, optional bool) (string, bool) {
	msg := utils.MESSAGE.Error.InvalidNbArgs
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

func (s *Server) showAllEvents() string {
	events := getEntitiesFromChannel(s.eChan, s)
	users := getEntitiesFromChannel(s.uChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)
	defer loadEntitiesToChannel(s.uChan, users, s)

	response := "Events:\n"

	for i := 1; i <= len(events); i++ {
		event := events[i]
		creator, _ := users[event.CreatorId]
		if event.Closed {
			response += "#" + strconv.Itoa(i) + ": " + event.Name + " (creator: " + creator.Username + ")" +
				utils.RED + " Closed" + utils.RESET + "\n"
		} else {
			response += "#" + strconv.Itoa(i) + ": " + event.Name + " (creator: " + creator.Username + ")" +
				utils.GREEN + " Open" + utils.RESET + "\n"
		}

	}
	response += "\n"

	return response
}

func (s *Server) showEvent(idEvent int) (string, bool) {
	events := getEntitiesFromChannel(s.eChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)

	event, ok := events[idEvent]

	if ok {
		users := getEntitiesFromChannel(s.uChan, s)
		defer loadEntitiesToChannel(s.uChan, users, s)
		creator, _ := users[event.CreatorId]
		response := "#" + strconv.Itoa(idEvent) + ": " + event.Name + " (creator: " + creator.Username + ")\n"
		response += "Jobs:\n"

		for i := 1; i <= len(event.Jobs); i++ {
			job := event.Jobs[i]
			if job.EventId == idEvent {
				response += "Job " + strconv.Itoa(i) + ": " + job.Name + " (" + strconv.Itoa(len(job.VolunteerIds)) + "/" + strconv.Itoa(job.NbVolunteers) + ")\n"
			}
		}
		response += "\n"

		return response, true
	}

	return utils.MESSAGE.Error.EventNotFound, false
}

// Functions of each command
func (s *Server) help(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.HELP, false); !ok {
		return msg
	}

	return utils.MESSAGE.Help
}

func (s *Server) createEvent(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.CREATE, true); !ok {
		return msg
	}

	username := args[len(args)-2]
	password := args[len(args)-1]

	userId, okUser := s.verifyUser(username, password)

	if !okUser {
		return utils.MESSAGE.Error.AccessDenied
	}

	var nbVolunteersPerJob []int
	var jobsName []string

	for i := 1; i < len(args)-utils.CREATE.MinOptArgs; i++ {
		if i%utils.CREATE.MinOptArgs == 0 {
			if nbVolunteer, err := strconv.Atoi(args[i]); err != nil || nbVolunteer < 0 {
				return utils.MESSAGE.Error.NbVolunteersInteger
			} else {
				nbVolunteersPerJob = append(nbVolunteersPerJob, nbVolunteer)
			}
		} else {
			jobsName = append(jobsName, args[i])
		}
	}

	events := getEntitiesFromChannel(s.eChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)

	eventId := len(events) + 1
	currentJobId := 1
	newJobs := map[int]types.Job{}
	for i := 0; i < len(jobsName); i++ {
		newJob := types.Job{
			Name:         jobsName[i],
			CreatorId:    nbVolunteersPerJob[i],
			EventId:      userId,
			NbVolunteers: eventId,
			VolunteerIds: []int{},
		}
		newJobs[currentJobId] = newJob
		currentJobId++
	}

	newEvent := types.Event{Name: args[0], CreatorId: userId, Jobs: newJobs}
	events[eventId] = newEvent

	return utils.MESSAGE.WrapSuccess("Event #" + strconv.Itoa(eventId) + " " + newEvent.Name + " and " + strconv.Itoa(len(newJobs)) + " job(s)" + " created\n")
}

func (s *Server) close(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.CLOSE, false); !ok {
		return msg
	}

	idEvent, errEvent := strconv.Atoi(args[0])
	username := args[1]
	password := args[2]

	if errEvent != nil {
		return utils.MESSAGE.Error.MustBeInteger
	}

	userId, okUser := s.verifyUser(username, password)
	if !okUser {
		return utils.MESSAGE.Error.AccessDenied
	}

	errMsg, ok := s.closeEvent(idEvent, userId)

	if !ok {
		return errMsg
	}

	return utils.MESSAGE.WrapSuccess("Event #" + strconv.Itoa(idEvent) + " is closed.\n")
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
		return utils.MESSAGE.Error.MustBeInteger
	}

	userId, okUser := s.verifyUser(username, password)
	if !okUser {
		return utils.MESSAGE.Error.AccessDenied
	}

	events := getEntitiesFromChannel(s.eChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)

	event, okEvent := events[idEvent]

	if !okEvent {
		return utils.MESSAGE.Error.EventNotFound
	} else if event.Closed {
		return utils.MESSAGE.Error.EventClosed
	} else {
		if event.CreatorId == userId {
			return utils.MESSAGE.Error.CreatorRegister
		}
	}

	msg, okJob := s.addUserToJob(&event, idJob, userId)

	if !okJob {
		return msg
	}
	return utils.MESSAGE.WrapSuccess("User registered in job #" + strconv.Itoa(idJob) + " for Event #" + strconv.Itoa(idEvent) + " " + event.Name + ".\n")
}

func (s *Server) show(args []string) string {

	if len(args) == utils.SHOW.MinOptArgs {
		idEvent, err := strconv.Atoi(args[0])
		if err != nil {
			return utils.MESSAGE.Error.MustBeInteger
		}
		msg, _ := s.showEvent(idEvent)
		return msg
	} else {
		return s.showAllEvents()
	}
}

func (s *Server) jobs(args []string) string {
	if msg, ok := s.checkNbArgs(args, &utils.JOBS, false); !ok {
		return msg
	}

	idEvent, errEvent := strconv.Atoi(args[0])
	if errEvent != nil {
		return utils.MESSAGE.Error.MustBeInteger
	}

	events := getEntitiesFromChannel(s.eChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)

	event, ok := events[idEvent]
	if !ok {
		return utils.MESSAGE.Error.EventNotFound
	}

	users := getEntitiesFromChannel(s.uChan, s)
	defer loadEntitiesToChannel(s.uChan, users, s)

	response := "#" + strconv.Itoa(idEvent) + " " + event.Name + ":\n"
	suffix := ""
	var allUsersWorking []string
	for jobId, job := range event.Jobs {
		response += suffix + "#" + strconv.Itoa(jobId) + " " + job.Name + " (" + strconv.Itoa(len(job.VolunteerIds)) + "/" + strconv.Itoa(job.NbVolunteers) + ")"
		suffix = " | "
		for _, userId := range job.VolunteerIds {
			allUsersWorking = append(allUsersWorking, users[userId].Username)
		}
	}
	response += "\n"

	for _, name := range allUsersWorking {
		response += name + "\n"
	}

	return response + "\n"
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

	switch name {
	case utils.HELP.Name:
		response = s.help(args)
	case utils.CREATE.Name:
		response = s.createEvent(args)
	case utils.CLOSE.Name:
		response = s.close(args)
	case utils.REGISTER.Name:
		response = s.register(args)
	case utils.SHOW.Name:
		response = s.show(args)
	case utils.JOBS.Name:
		response = s.jobs(args)
	case utils.QUIT.Name:
		end = true
	default:
		response = utils.MESSAGE.Error.InvalidCommand
	}

	return response, end
}

func (s *Server) handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		input, err := reader.ReadString('\n')

		if err != nil {
			log.Println(utils.RED + "(ERROR) " + err.Error() + utils.RESET)
			break
		}

		if !s.Config.Silent {
			log.Print(utils.YELLOW + "(INFO) " + conn.RemoteAddr().String() + " -> " + strings.TrimSuffix(input, "\n") + utils.RESET)
		}
		response, end := s.processCommand(strings.TrimSpace(string(input)))

		if end {
			if !s.Config.Silent {
				log.Println(utils.RED + "(INFO) " + conn.RemoteAddr().String() + " disconnected" + utils.RESET)
			}
			break
		}

		conn.Write([]byte(response))
	}
	conn.Close()
}

func (s *Server) Run() {

	users, events := utils.GetEntities(entities)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(s.Config.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	s.eChan = make(chan map[int]types.Event, 1)
	s.uChan = make(chan map[int]types.User, 1)

	s.uChan <- users
	s.eChan <- events

	if !s.Config.Silent {
		log.Println(utils.GREEN + "(INFO) " + "Server started on port " + strconv.Itoa(s.Config.Port) + utils.RESET)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(utils.RED + "(ERROR) " + err.Error() + utils.RESET)
			return
		} else {
			if !s.Config.Silent {
				log.Println(utils.GREEN + "(INFO) " + conn.RemoteAddr().String() + " connected" + utils.RESET)
			}
		}
		go s.handleConn(conn)
	}
}
