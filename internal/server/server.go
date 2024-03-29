// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 2 SDR

// Package server propose un serveur TCP qui effectue une gestion de manifestations.
//
// Le serveur est capable de gérer plusieurs clients en même temps.
// Le serveur est capable de se connecter à d'autres serveurs pour former un réseau et gère les accès à une section critique
// en utilisant l'algorithme de Lamport optimisé.
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
var entities string                             // variable qui permet de charger le fichier des entités dans les binaries finales de l'application
var users, events = utils.GetEntities(entities) // charge les utilisateurs et les événements depuis le fichier entities.json

// Channels utilisés pour la réception des inputs et actions liées aux clients
var inputChan = make(chan string, 1) // channel récupérant l'entrée d'un client connecté
var resChan = make(chan string, 1)   // channel stockant la réponse du serveur à un input d'un client
var quitChan = make(chan bool, 1)    // channel permettant de terminer une session d'un client

// Channels utilisés pour traiter les communications pour l'algorithme de Lamport dans la goroutine principale
var reqChan = make(chan bool, 1)                 // Envoi d'un REQ aux autres serveurs
var accessChan = make(chan bool, 1)              // Accès à la section critique de l'algorithme de Lamport
var relChan = make(chan bool, 1)                 // Envoi d'un REL aux autres serveurs
var commChan = make(chan types.Communication, 1) // Réception des communications des autres serveurs (REQ, REL, ACK)

var hasAccess = false // Booléen représentant la possession de la section critique de l'algorithme de Lamport

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
// Chaque connexion est ensuite gérée par plusieurs goroutines jusqu'à sa fermeture.
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

	// Boucle acceptant les connexions des clients
	for {
		conn, err := clientListener.Accept()
		if err != nil {
			s.log(types.ERROR, err.Error())
		} else {
			reader := bufio.NewReader(conn)

			// Récupère le nom du client
			nameStr, err := reader.ReadString('\n')
			if err != nil {
				s.log(types.ERROR, err.Error())
			}

			name := strings.TrimSuffix(nameStr, "\n")
			s.log(types.INFO, utils.GREEN+name+" connected"+utils.RESET)

			go s.handleClientConns(conn, name)
		}
	}
}

// initServersConns initialise les connexions avec les autres serveurs.
// La méthode s'assure que le serveur ait une connexion (en tant que client ou serveur) avec tous les autres serveurs
// présents dans sa configuration. Pour cela, s'il n'arrive pas à se connecter à un serveur, il passe en mode "server" et
// attend que les autres serveurs se connectent à lui.
// Finalement, il lance une goroutine qui va traiter les communications entre les serveurs.
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

	// Lance la goroutine exécutant la boucle principale de l'algorithme de Lamport
	go func() {
		for {
			select {
			case <-reqChan: // Demande d'accès à la section critique
				if len(s.Config.Servers) == 1 {
					accessChan <- true
					continue
				}
				s.Stamp++
				s.sendComm(types.Request, utils.MapKeysToArray(s.conns), nil)
			case <-accessChan: // Permission de l'accès à la section critique
				hasAccess = true
			case <-relChan: // Libération de la section critique
				hasAccess = false
				s.Stamp++
				s.sendComm(types.Release, utils.MapKeysToArray(s.conns), &events)
			case comm := <-commChan: // Traitement d'une communication reçue
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

// handleIncomingComms gère les communications entrantes des autres serveurs.
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

	if err := conn.Close(); err != nil {
		s.log(types.ERROR, err.Error())
	}
}

// ---------- Méthodes concernant les communications serveurs-serveurs & Lamport ----------

// verifyCriticalSection vérifie si le serveur peut accéder à la section critique distribuée selon l'algorithme de Lamport.
func (s *Server) verifyCriticalSection() {
	if hasAccess {
		return
	}

	if s.comms[s.Number].Type != types.Request {
		return
	}

	hasOldestReq := true
	for i := 1; i <= len(s.Config.Servers); i++ {
		if i == s.Number {
			continue
		}
		if s.comms[s.Number].Stamp > s.comms[i].Stamp || (s.comms[s.Number].Stamp == s.comms[i].Stamp && s.Number > s.comms[i].From) {
			hasOldestReq = false
			break
		}
	}
	if hasOldestReq {
		accessChan <- true
	}
}

// sendComm prépare et envoie une communication à un ou plusieurs serveurs. Cette communication est envoyée en JSON et
// est stockée dans la map des communications du serveur à son propre index. La méthode peut prendre la map des
// manifestations (dans le cas d'un REL par exemple) pour communiquer aux autres serveurs la version à jour de l'entité.
func (s *Server) sendComm(commType types.CommunicationType, to []int, payload *map[int]types.Event) {
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

	s.comms[s.Number] = communication
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

// handleRequest gère la réception d'une requête (REQ) d'accès à la section critique distribuée. Avec Lamport optimisée,
// si le serveur actuel contient déjà une requête, il ne fait rien. Sinon, il envoie un ACK au serveur qui a envoyé la REQ.
// Dans les deux cas, le serveur vérifie si il peut accéder à la section critique.
func (s *Server) handleRequest(comm types.Communication) {

	s.Stamp = utils.Max(s.Stamp, comm.Stamp) + 1
	s.comms[comm.From] = comm
	s.log(types.LAMPORT, "STATUS: "+s.commsToString()+" IN  "+string(comm.Type)+strconv.Itoa(comm.Stamp)+" FROM S"+strconv.Itoa(comm.From))

	if s.comms[s.Number].Type != types.Request {
		s.sendComm(types.Acknowledge, []int{comm.From}, nil)
	}

	s.verifyCriticalSection()
}

// handleAcknowledge gère la réception d'un ACK d'accès à la section critique distribuée. Si le serveur actuel contient
// déjà une requête, il s'occupe seulement de vérifier s'il a accès à la section critique.
func (s *Server) handleAcknowledge(comm types.Communication) {
	s.Stamp = utils.Max(s.Stamp, comm.Stamp) + 1
	if s.comms[comm.From].Type != types.Request {
		s.comms[comm.From] = comm
		s.log(types.LAMPORT, "STATUS: "+s.commsToString()+" IN  "+string(comm.Type)+strconv.Itoa(comm.Stamp)+" FROM S"+strconv.Itoa(comm.From))
	}

	s.verifyCriticalSection()
}

// handleRelease gère la réception d'un REL d'accès à la section critique distribuée. Si le REL contient un payload,
// le serveur met à jour sa map des manifestations. Finalement, le serveur vérifie s'il a accès à la section critique.
func (s *Server) handleRelease(comm types.Communication) {
	s.Stamp = utils.Max(s.Stamp, comm.Stamp) + 1
	s.comms[comm.From] = comm
	events = comm.Payload
	s.log(types.LAMPORT, "STATUS: "+s.commsToString()+" IN  "+string(comm.Type)+strconv.Itoa(comm.Stamp)+" FROM S"+strconv.Itoa(comm.From))

	s.verifyCriticalSection()
}

// ---------- Méthodes pour la gestion des clients et leurs commandes ----------

// handleClientConns gère l'I/O avec un client connecté au serveur
func (s *Server) handleClientConns(conn net.Conn, name string) {
	reader := bufio.NewReader(conn)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			s.log(types.ERROR, err.Error())
			break
		}

		s.log(types.INFO, utils.YELLOW+name+" -> "+strings.TrimSuffix(input, "\n")+utils.RESET)
		inputChan <- input

		select {
		case response := <-resChan:
			_, err := conn.Write([]byte(response))
			if err != nil {
				s.log(types.ERROR, err.Error())
			}
		case <-quitChan:
			s.log(types.INFO, utils.RED+name+" disconnected"+utils.RESET)
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

	s.debugTrace(true)
	reqChan <- true
	<-accessChan
	s.log(types.LAMPORT, utils.GREEN+"ACCESSING DISTRIBUTED CRITICAL SECTION"+utils.RESET)

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

	s.log(types.LAMPORT, utils.RED+"RELEASING DISTRIBUTED CRITICAL SECTION"+utils.RESET)
	resChan <- response
	relChan <- true
	s.debugTrace(false)
}

// ---------- Méthode pour chaque commande ----------

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

// ---------- Méthodes helpers ----------

// commsToString affiche la map des communications du serveur en un tableau de string.
func (s *Server) commsToString() string {
	var str string
	str += "["
	for i := 1; i <= len(s.Config.Servers); i++ {
		str += "S" + strconv.Itoa(i) + ": " + string(s.comms[i].Type) + strconv.Itoa(s.comms[i].Stamp)
		if i != len(s.Config.Servers) {
			str += ", "
		}
	}
	str += "]"

	return str
}

// debugTrace permet d'afficher des informations de debugTrace si le mode debugTrace est activé.
//
// La méthode ralentit artificiellement l'exécution du serveur pour tester les accès concurrents d'une durée égale à la propriété
// DebugDelay de Config. Le paramètre start indique s'il s'agit d'un début ou d'une fin d'accès à une section critique.
func (s *Server) debugTrace(start bool) {
	if s.Config.Debug {
		if start {
			s.log(types.DEBUG, utils.RED+"ACCESSING LOCAL CRITICAL SECTION"+utils.RESET)
			time.Sleep(time.Duration(s.Config.DebugDelay) * time.Second)
		} else {
			s.log(types.DEBUG, utils.GREEN+"RELEASING LOCAL CRITICAL SECTION"+utils.RESET)
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
