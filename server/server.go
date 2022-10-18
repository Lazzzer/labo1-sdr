// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 1 SDR

// Package server propose un serveur TCP qui effectue une gestion de manifestations.
//
// Le serveur est capable de gérer plusieurs clients en même temps.
// Au démarrage, le serveur charge une configuration depuis un fichier config.json.
// Il charge ensuite les utilisateurs et les événements depuis un fichier entities.json.
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
var entities string // variable qui permet de charger le fichier des entités dans les binaries finales de l'application

// Server est une structure permettant de gérer un serveur TCP.
type Server struct {
	Config types.Config             // Configuration du serveur
	eChan  chan map[int]types.Event // Canal d'accès à la map contenant des manifestations
	uChan  chan map[int]types.User  // Canal d'accès à la map contenant des utilisateurs
}

// ---------- Fonctions helpers ----------

// getEntitiesFromChannel permet de récupérer une map de manifestations ou d'utilisateurs depuis un canal.
//
// En mode debug, le serveur attend un certain laps de temps et log l'accès à la section critique en question avant de retourner l'entité.
func getEntitiesFromChannel[T types.Event | types.User](ch <-chan map[int]T, s *Server) map[int]T {
	entities := <-ch
	s.debug(reflect.TypeOf(&entities).Elem().Elem().String(), true)

	return entities
}

// loadEntitiesToChannel permet de charger une map de manifestations ou d'utilisateurs dans un canal.
//
// En mode debug, le serveur attend un certain laps de temps et log l'accès à la section critique en question avant de laisser l'exécution
// se poursuivre.
func loadEntitiesToChannel[T types.Event | types.User](ch chan<- map[int]T, entities map[int]T, s *Server) {
	ch <- entities
	s.debug(reflect.TypeOf(&entities).Elem().Elem().String(), false)
}

// debug permet d'afficher des informations de debug si le mode debug est activé.
//
// La méthode ralentit artificiellement l'exécution du serveur pour tester les accès concurrents d'une durée égale à la propriété
// DebugDelay de Config. Le paramètre start indique s'il s'agit d'un début ou d'une fin d'accès à une section critique.
func (s *Server) debug(entity string, start bool) {
	if s.Config.Debug {
		if start {
			log.Println(utils.ORANGE + "(DEBUG) " + utils.RED + "ACCESSING" + utils.ORANGE + " shared section for entity: " + entity + utils.RESET)
			time.Sleep(time.Duration(s.Config.DebugDelay) * time.Second)
		} else {
			log.Println(utils.ORANGE + "(DEBUG) " + utils.GREEN + "RELEASING" + utils.ORANGE + " shared section for entity: " + entity + utils.RESET)
		}
	}
}

// verifyUser permet de vérifier si un utilisateur existe dans la map des utilisateurs et retourne sa clé dans la map et un booléen
// indiquant sa présence.
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

// removeUserInJob permet de supprimer l'id d'un utilisateur du tableau des utilisateurs qui ont postulé à un job et retourne si l'opération
// a réussi.
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

// addUserToJob permet d'ajouter un utilisateur à un job et retourne un message vide et true si l'opération a réussi.
// En cas d'échec d'ajout, la méthode retourne un message d'erreur spécifique et false.
//
// Si un utilisateur est déjà dans un job de la même manifestation, sa postulation est supprimée et il est ajouté dans le nouveau job.
func (s *Server) addUserToJob(event *types.Event, idJob, idUser int) (string, bool) {

	job, ok := event.Jobs[idJob]

	if ok {
		// Différentes vérifications selon le cahier des charges avec les messages d'erreur correspondants
		if event.CreatorId == idUser {
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

// closeEvent permet de fermer une manifestation et retourne un message vide et true si l'opération a réussi.
// En cas d'échec de fermeture, la méthode retourne un message d'erreur spécifique et false.
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

// checkNbArgs permet de vérifier le nombre d'arguments d'une commande et retourne un message vide et true si le nombre d'arguments est correct.
// En cas d'échec de vérification, la méthode retourne un message d'erreur spécifique et false.
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

// showAllEvents permet d'afficher toutes les manifestations.
func (s *Server) showAllEvents() string {
	events := getEntitiesFromChannel(s.eChan, s)
	users := getEntitiesFromChannel(s.uChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)
	defer loadEntitiesToChannel(s.uChan, users, s)

	response := "Events:\n"

	for i := 1; i <= len(events); i++ {
		event := events[i]
		creator := users[event.CreatorId]
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

// showEvent permet d'afficher la manifestation correspondant à l'identifiant passé en paramètre et retourne un message vide et true
// si l'opération a réussi. En cas d'échec d'affichage, la méthode retourne un message d'erreur spécifique et false.
func (s *Server) showEvent(idEvent int) (string, bool) {
	events := getEntitiesFromChannel(s.eChan, s)
	defer loadEntitiesToChannel(s.eChan, events, s)

	event, ok := events[idEvent]

	if ok {
		users := getEntitiesFromChannel(s.uChan, s)
		defer loadEntitiesToChannel(s.uChan, users, s)
		creator := users[event.CreatorId]
		response := "#" + strconv.Itoa(idEvent) + ": " + event.Name + " (creator: " + creator.Username + ")\n"
		response += "Jobs:\n"

		for i := 1; i <= len(event.Jobs); i++ {
			job := event.Jobs[i]
			response += "Job " + strconv.Itoa(i) + ": " + job.Name + " (" + strconv.Itoa(len(job.VolunteerIds)) + "/" + strconv.Itoa(job.NbVolunteers) + ")\n"
		}

		response += "\n"

		return response, true
	}

	return utils.MESSAGE.Error.EventNotFound, false
}

// ---------- Fonctions pour chaque commande ----------

// help est la méthode appelée par la commande "help" et affiche un message d'aide listant chaque commande et ses arguments.
func (s *Server) help(args []string) string {

	if msg, ok := s.checkNbArgs(args, &utils.HELP, false); !ok {
		return msg
	}

	return utils.MESSAGE.Help
}

// createEvent est la méthode appelée par la commande "create" et  permet de créer une manifestation et retourne un message de confirmation.
// En cas d'échec de création, la méthode retourne un message d'erreur spécifique.
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
			NbVolunteers: nbVolunteersPerJob[i],
			VolunteerIds: []int{},
		}
		newJobs[currentJobId] = newJob
		currentJobId++
	}

	newEvent := types.Event{Name: args[0], CreatorId: userId, Jobs: newJobs}
	events[eventId] = newEvent

	return utils.MESSAGE.WrapSuccess("Event #" + strconv.Itoa(eventId) + " " + newEvent.Name + " and " + strconv.Itoa(len(newJobs)) + " job(s)" + " created\n")
}

// closeEvent est la méthode appelée par la commande "close" et permet de fermer une manifestation et retourne un message de confirmation.
// En cas d'échec de fermeture, la méthode retourne un message d'erreur spécifique.
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

// register est la méthode appelée par la commande "register" et permet d'inscrire un utilisateur à un job d'une manifestation et retourne un message de confirmation.
// En cas d'échec d'inscription, la méthode retourne un message d'erreur spécifique.
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

// show est la méthode appelée par la commande "show" et permet d'afficher les manifestations et leurs informations.
// En passant un identifiant de manifestation en argument dans la commande, la méthode affiche les informations de la manifestation avec ses jobs.
func (s *Server) show(args []string) string {
	if len(args) == utils.SHOW.MinOptArgs {
		idEvent, err := strconv.Atoi(args[0])
		if err != nil {
			return utils.MESSAGE.Error.MustBeInteger
		}
		msg, _ := s.showEvent(idEvent)
		return msg
	} else if len(args) == 0 {
		return s.showAllEvents()
	} else {
		return utils.MESSAGE.Error.InvalidCommand
	}
}

// jobs est la méthode appelée par la commande "jobs" et permet d'afficher la répartition des bénévoles et des jobs d'une manifestation.
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
	for i := 1; i <= len(event.Jobs); i++ {
		job := event.Jobs[i]
		response += suffix + "#" + strconv.Itoa(i) + " " + job.Name + " (" + strconv.Itoa(len(job.VolunteerIds)) + "/" + strconv.Itoa(job.NbVolunteers) + ")"
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

// processCommand permet de traiter l'entrée utilisateur et de lancer la méthode correspondante à la commande saisie.
// La méthode notifie au serveur l'arrêt de sa boucle de traitement des commandes lorsque la commande "quit" est saisie.
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

// handleConn gère l'I/O avec un client connecté au serveur
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

// Run lance le serveur et attend les connexions des clients.
//
// Chaque connexion est ensuite gérée par une goroutine jusqu'à sa fermeture.
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
