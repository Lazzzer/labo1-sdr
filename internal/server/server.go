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
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Lazzzer/labo1-sdr/internal/utils"
	"github.com/Lazzzer/labo1-sdr/internal/utils/types"
)

//go:embed entities.json
var entities string // variable qui permet de charger le fichier des entités dans les binaries finales de l'application
var users, events = utils.GetEntities(entities)
var inputChan = make(chan string, 1)
var resChan = make(chan string, 1)
var quitChan = make(chan bool, 1)
var commChan = make(chan types.Communication, 1)
var commsAccessChan = make(chan map[int]types.Communication, 1)
var reqChan = make(chan bool, 1)
var relChan = make(chan bool, 1)
var hasSectionChan = make(chan bool, 1)
var hasAccess = false
var accessChan = make(chan bool, 1)

// Server est une struct représentant un serveur TCP.
type Server struct {
	Number     int                         // numéro du serveur
	Port       string                      // port sur lequel le serveur écoute
	ClientPort string                      // port sur lequel le serveur écoute les connexions des clients
	Config     types.ServerConfig          // Configuration du serveur
	Stamp      int                         // Estampille actuelle du serveur
	conns      map[int]net.Conn            // Map de connexions des serveurs
	comms      map[int]types.Communication // Map des dernières communications entre les serveurs
}

// Run lance le serveur et attend les connexions des clients.
//
// Chaque connexion est ensuite gérée par une goroutine jusqu'à sa fermeture.
func (s *Server) Run() {

	var err error
	var srvListener net.Listener
	var clientListener net.Listener

	srvListener, err = net.Listen("tcp", ":"+s.Port)
	if err != nil {
		log.Fatal(err)
	}

	s.initServersConns(srvListener)

	// Le serveur est prêt à recevoir des connexions de clients
	s.log(types.INFO, "Listening for clients connections on port "+s.ClientPort)
	clientListener, err = net.Listen("tcp", ":"+s.ClientPort)
	if err != nil {
		log.Fatal(err)
	}

	// Traite les inputs des différents clients un par un
	go func() {
		for {
			input := <-inputChan
			s.processCommand(input)
		}
	}()

	for {
		conn, err := clientListener.Accept()
		if err != nil {
			s.log(types.ERROR, err.Error())
			break // TODO: Better error handling ?
		} else {
			s.log(types.INFO, utils.GREEN+conn.RemoteAddr().String()+" connected"+utils.RESET)
		}
		go s.handleClientConn(conn)
	}

	err = srvListener.Close()
	if err != nil {
		s.log(types.ERROR, err.Error())
	}

	err = clientListener.Close()
	if err != nil {
		s.log(types.ERROR, err.Error())
	}
}

func (s *Server) initServersConns(listener net.Listener) {
	s.conns = make(map[int]net.Conn, len(s.Config.Servers)-1)
	nbSuccessConn := 0

	// Se connecte à chaque serveur déjà en ligne
	for number := 1; number <= len(s.Config.Servers); number++ {
		if number != s.Number {
			conn, err := net.Dial("tcp", s.Config.Servers[number])
			if err != nil {
				s.log(types.INFO, utils.RED+"Server #"+strconv.Itoa(s.Number)+" could not connect to Server #"+strconv.Itoa(number)+utils.RESET)
				continue
			} else {
				// Ajout de la connexion établie dans la map
				s.log(types.INFO, utils.GREEN+"Server #"+strconv.Itoa(s.Number)+" connected to Server #"+strconv.Itoa(number)+utils.RESET)
				s.conns[number] = conn
				nbSuccessConn++
				// Envoi de son numéro au serveur qui écoute sa connexion
				_, err = conn.Write([]byte(strconv.Itoa(s.Number) + "\n"))
				if err != nil {
					s.log(types.ERROR, err.Error())
				}
			}
		}
	}

	// Se met en mode attente de connexion des autres serveurs s'il n'arrive plus à se connecter à un serveur
	if nbSuccessConn < len(s.Config.Servers)-1 {
		s.log(types.INFO, "Listening for missing servers connections")
		for nbSuccessConn < len(s.Config.Servers)-1 {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			} else {
				s.handleHandshake(conn)
				nbSuccessConn++
			}
		}
	}

	// Initialise l'estampille et la map des communications avec des REL0
	s.Stamp = 0
	s.comms = make(map[int]types.Communication, len(s.Config.Servers))

	for number := 1; number <= len(s.Config.Servers); number++ {
		s.comms[number] = types.Communication{
			Type:  types.Release,
			Stamp: 0,
		}
	}

	commsAccessChan <- s.comms

	go func() {
		for {
			select {
			case <-reqChan: // Demande d'accès à la section critique
				log.Println("ACCESS reqChan")
				s.Stamp++
				s.sendComm(types.Request, utils.MapKeysToArray(s.conns), nil)
			case <-hasSectionChan:
				hasAccess = true
				accessChan <- hasAccess
			case <-relChan: // Libération de la section critique
				log.Println("ACCESS relChan")
				hasAccess = false
				s.Stamp++
				s.sendComm(types.Release, utils.MapKeysToArray(s.conns), &events)
			case comm := <-commChan: // Réception d'une communication
				switch comm.Type {
				case types.Request:
					s.handleRequest(comm)
				case types.Acknowledge:
					s.handleAcknowledge(comm)
				case types.Release:
					s.handleRelease(comm)
				}
			}
		}
	}()

	// Lance une goroutine pour chaque serveur connecté qui gère les communications entrantes de synchronisation de l'algorithme de Lamport
	for _, conn := range s.conns {
		go s.handleIncomingComms(conn)
	}
}

// handleHandshake gère la première communication d'un serveur qui reçoit la connexion d'un autre serveur. Cette méthode sert surtout
// à récupérer le numéro du serveur "client" pour pouvoir l'ajouter à la liste des connexions du serveur.
func (s *Server) handleHandshake(conn net.Conn) {
	reader := bufio.NewReader(conn)

	// Récupère le numéro du serveur
	numberStr, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	number, err := strconv.Atoi(strings.TrimSuffix(numberStr, "\n"))
	if err != nil {
		log.Fatal(err)
	}

	s.log(types.INFO, utils.GREEN+"Server #"+strconv.Itoa(s.Number)+" received a connection from Server #"+strings.TrimSuffix(numberStr, "\n")+utils.RESET)
	s.conns[number] = conn
}

func (s *Server) handleIncomingComms(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			s.log(types.ERROR, err.Error())
			break
		}

		var comm types.Communication
		err = json.Unmarshal([]byte(strings.TrimSuffix(input, "\n")), &comm)
		if err != nil {
			s.log(types.ERROR, err.Error())
			break
		}

		commChan <- comm
	}
}

func (s *Server) commsToString() string {
	comms := <-commsAccessChan
	defer func() { commsAccessChan <- comms }()

	var str string
	str += "["
	for i := 1; i <= len(s.Config.Servers); i++ {
		str += "S" + strconv.Itoa(i) + ": " + string(comms[i].Type) + strconv.Itoa(comms[i].Stamp)
		if i != len(s.Config.Servers) {
			str += ", "
		}
	}
	str += "]"

	return str
}
func (s *Server) verifyCriticalSection() {
	if hasAccess {
		return
	}

	comms := <-commsAccessChan

	if comms[s.Number].Type != types.Request {
		commsAccessChan <- comms
		return
	}

	hasOldestReq := true
	for i := 1; i <= len(s.Config.Servers); i++ {
		if i == s.Number {
			continue
		}
		if comms[s.Number].Stamp > comms[i].Stamp || (comms[s.Number].Stamp == comms[i].Stamp && s.Number > comms[i].From) {
			hasOldestReq = false
			break
		}
	}
	commsAccessChan <- comms
	if hasOldestReq {
		hasSectionChan <- true
		s.log(types.LAMPORT, utils.GREEN+"Server #"+strconv.Itoa(s.Number)+" has access to the critical section"+utils.RESET)
	}

}

func (s *Server) sendComm(commType types.CommunicationType, to []int, payload *map[int]types.Event) {
	comms := <-commsAccessChan

	communication := types.Communication{
		Type:  commType,
		From:  s.Number,
		To:    to,
		Stamp: s.Stamp,
	}
	if payload == nil {
		communication.Payload = nil
	} else {
		communication.Payload = *payload
	}

	comms[s.Number] = communication
	commsAccessChan <- comms
	s.log(types.LAMPORT, "STATUS: "+s.commsToString()+" OUT "+string(communication.Type)+strconv.Itoa(communication.Stamp)+" TO "+utils.IntToString(communication.To))

	communicationJson, err := json.Marshal(communication)
	if err != nil {
		s.log(types.ERROR, err.Error())
	}

	for _, number := range to {
		_, err := s.conns[number].Write([]byte(string(communicationJson) + "\n"))
		if err != nil {
			s.log(types.ERROR, err.Error())
		}
	}
}

func (s *Server) handleRequest(comm types.Communication) {
	comms := <-commsAccessChan

	s.Stamp = utils.Max(s.Stamp, comm.Stamp) + 1
	comms[comm.From] = comm
	commsAccessChan <- comms
	s.log(types.LAMPORT, "STATUS: "+s.commsToString()+" IN  "+string(comm.Type)+strconv.Itoa(comm.Stamp)+" FROM S"+strconv.Itoa(comm.From))

	if comms[s.Number].Type != types.Request {
		s.sendComm(types.Acknowledge, []int{comm.From}, nil)
	}

	s.verifyCriticalSection()
}

func (s *Server) handleAcknowledge(comm types.Communication) {
	comms := <-commsAccessChan

	s.Stamp = utils.Max(s.Stamp, comm.Stamp) + 1
	if comms[comm.From].Type != types.Request {
		comms[comm.From] = comm
		commsAccessChan <- comms
		s.log(types.LAMPORT, "STATUS: "+s.commsToString()+" IN  "+string(comm.Type)+strconv.Itoa(comm.Stamp)+" FROM S"+strconv.Itoa(comm.From))
	}

	s.verifyCriticalSection()
}

func (s *Server) handleRelease(comm types.Communication) {
	comms := <-commsAccessChan

	s.Stamp = utils.Max(s.Stamp, comm.Stamp) + 1
	comms[comm.From] = comm
	events = comm.Payload
	commsAccessChan <- comms
	s.log(types.LAMPORT, "STATUS: "+s.commsToString()+" IN  "+string(comm.Type)+strconv.Itoa(comm.Stamp)+" FROM S"+strconv.Itoa(comm.From))

	s.verifyCriticalSection()
}

// ---------- Fonctions pour la gestion des clients ----------

// handleClientConn gère l'I/O avec un client connecté au serveur
func (s *Server) handleClientConn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			s.log(types.ERROR, err.Error())
			break
		}

		s.log(types.INFO, utils.YELLOW+conn.RemoteAddr().String()+" -> "+strings.TrimSuffix(input, "\n")+utils.RESET)
		inputChan <- input

		select {
		case response := <-resChan:
			_, err := conn.Write([]byte(response))
			if err != nil {
				s.log(types.ERROR, err.Error())
			}
		case <-quitChan:
			s.log(types.INFO, utils.RED+conn.RemoteAddr().String()+" disconnected"+utils.RESET)
			err := conn.Close()
			if err != nil {
				s.log(types.ERROR, err.Error())
			}
			return
		}
	}
}

// processCommand permet de traiter l'entrée utilisateur et de lancer la méthode correspondante à la commande saisie.
// La méthode notifie au serveur l'arrêt de sa boucle de traitement des commandes lorsque la commande "quit" est saisie.
func (s *Server) processCommand(input string) {
	args := strings.Fields(input)

	if len(args) == 0 {
		resChan <- "Empty command"
		return
	}

	name := args[0]
	args = args[1:]

	// Commandes n'ayant pas besoin d'accès à la section critique

	switch name {
	case utils.QUIT.Name:
		quitChan <- true
		return
	case utils.HELP.Name:
		resChan <- s.help(args)
		return
	}

	reqChan <- true
	<-accessChan
	log.Println("ACCESS accessChan")

	s.debugTrace(true)

	var response string

	// Commandes avec accès à la section critique
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
	default:
		response = utils.MESSAGE.Error.InvalidCommand
	}

	resChan <- response
	relChan <- true
	s.debugTrace(false)
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

	event, ok := events[idEvent]
	if !ok {
		return utils.MESSAGE.Error.EventNotFound
	}

	eventTitle := "#" + strconv.Itoa(idEvent) + " " + utils.BOLD + utils.CYAN + event.Name + utils.RESET + "\n\n"
	firstLine := utils.BOLD + "Volunteers" + utils.RESET + "\t"
	numberOfUsers := 0
	allUsersWorking := make([][]string, len(event.Jobs))
	for i := 1; i <= len(event.Jobs); i++ {
		job := event.Jobs[i]
		firstLine += "#" + strconv.Itoa(i) + " " + job.Name + " (" + strconv.Itoa(len(job.VolunteerIds)) + "/" + strconv.Itoa(job.NbVolunteers) + ")\t"
		for _, userId := range job.VolunteerIds {
			allUsersWorking[i-1] = append(allUsersWorking[i-1], users[userId].Username)
			numberOfUsers++
		}
	}

	var builder strings.Builder
	aligner := "\t"
	var endColumn string
	for i := 0; i < len(allUsersWorking); i++ {
		endColumn += "\t"
	}

	w := tabwriter.NewWriter(&builder, 0, 0, 3, ' ', 0)
	_, err := fmt.Fprintln(w, firstLine)
	if err != nil {
		s.log(types.ERROR, err.Error())
	}

	if numberOfUsers == 0 {
		err := w.Flush()
		if err != nil {
			s.log(types.ERROR, err.Error())
		}
		return utils.MESSAGE.WrapEvent(eventTitle + builder.String() + "\nThere is currently no volunteers for this event.\n")
	}

	for i := 0; i < len(allUsersWorking); i++ {
		for j := 0; j < len(allUsersWorking[i]); j++ {
			_, err = fmt.Fprintln(w, allUsersWorking[i][j]+aligner+"✅"+endColumn)
			if err != nil {
				s.log(types.ERROR, err.Error())
			}
		}
		aligner += "\t"
		endColumn = strings.TrimSuffix(endColumn, "\t")
	}
	err = w.Flush()
	if err != nil {
		s.log(types.ERROR, err.Error())
	}

	return utils.MESSAGE.WrapEvent(eventTitle + builder.String())
}

// ---------- Fonctions helpers ----------

// debugTrace permet d'afficher des informations de debugTrace si le mode debugTrace est activé.
//
// La méthode ralentit artificiellement l'exécution du serveur pour tester les accès concurrents d'une durée égale à la propriété
// DebugDelay de Config. Le paramètre start indique s'il s'agit d'un début ou d'une fin d'accès à une section critique.
func (s *Server) debugTrace(start bool) {
	if s.Config.Debug {
		if start {
			s.log(types.DEBUG, utils.RED+"ACCESSING SHARED SECTION"+utils.RESET)
			time.Sleep(time.Duration(s.Config.DebugDelay) * time.Second)
		} else {
			s.log(types.DEBUG, utils.GREEN+"RELEASING SHARED SECTION"+utils.RESET)
		}
	}
}

func (s *Server) log(logType types.LogType, message string) {
	if !s.Config.Silent {
		switch logType {
		case types.INFO:
			log.Println(utils.CYAN + "(INFO) " + utils.RESET + message)
		case types.ERROR:
			log.Println(utils.RED + "(ERROR) " + utils.RESET + message)
		case types.DEBUG:
			log.Println(utils.ORANGE + "(DEBUG) " + utils.RESET + message)
		case types.LAMPORT:
			log.Println(utils.PINK + "(LAMPORT) " + utils.RESET + message)

		}
	}
}

// verifyUser permet de vérifier si un utilisateur existe dans la map des utilisateurs et retourne sa clé dans la map et un booléen
// indiquant sa présence.
func (s *Server) verifyUser(username, password string) (int, bool) {

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
	var response string

	for i := 1; i <= len(events); i++ {
		event := events[i]
		creator := users[event.CreatorId]
		if event.Closed {
			response += utils.RED + "Closed" + utils.RESET
		} else {
			response += utils.GREEN + "Open" + utils.RESET
		}
		response += "\t#" + strconv.Itoa(i) + " " + utils.BOLD + utils.CYAN + event.Name + utils.RESET + " / Creator: " + creator.Username + "\n"
		if i != len(events) {
			response += "\n"
		}
	}

	return utils.MESSAGE.WrapEvent(response)
}

// showEvent permet d'afficher la manifestation correspondant à l'identifiant passé en paramètre et retourne un message vide et true
// si l'opération a réussi. En cas d'échec d'affichage, la méthode retourne un message d'erreur spécifique et false.
func (s *Server) showEvent(idEvent int) (string, bool) {
	event, ok := events[idEvent]

	if ok {
		creator := users[event.CreatorId]

		response := "#" + strconv.Itoa(idEvent) + " " + utils.BOLD + utils.CYAN + event.Name + utils.RESET + "\n\n"
		response += "Creator: " + creator.Username + "\n\n"
		response += "🦺" + utils.BOLD + " Jobs" + utils.RESET + "\n\n"

		for i := 1; i <= len(event.Jobs); i++ {
			job := event.Jobs[i]

			var color string
			if len(job.VolunteerIds) == job.NbVolunteers {
				color = utils.RED
			} else {
				color = utils.GREEN
			}

			response += color + "(" + strconv.Itoa(len(job.VolunteerIds)) + "/" + strconv.Itoa(job.NbVolunteers) + ")" + utils.RESET + "\tJob #" + strconv.Itoa(i) + ": " + job.Name + "\n"
		}

		return utils.MESSAGE.WrapEvent(response), true
	}

	return utils.MESSAGE.Error.EventNotFound, false
}
