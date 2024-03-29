// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 2 SDR

// Package types propose différents types utilisés par l'application pour parser les fichiers de configuration et les entités.
package types

// Config représente la configuration partagée par un serveur et un client.
// Elle contient la liste des adresses des serveurs ainsi que leur numéro.
// Pour le client, Address représente l'adresse du serveur auquel il se connecte.
// Pour le serveur, Address représente l'adresse sur laquelle il écoute.
type Config struct {
	Address string         `json:"adress,omitempty"` // Adresse du serveur
	Servers map[int]string `json:"servers"`          // Adresses des serveurs disponibles
}

// ServerConfig est une configuration de serveur contenant notamment la valeur des flags et la liste des ports à utiliser
// et à exposer pour l'écoute des connexions clientes.
type ServerConfig struct {
	Config
	ClientPorts map[int]string `json:"client_ports"`          // Ports des listeners à utiliser pour écouter les connexions des clients
	Debug       bool           `json:"debug"`                 // Activation du mode debug pour vérifier la concurrence
	Silent      bool           `json:"silent"`                // Activation du mode silencieux pour ne pas afficher les logs
	DebugDelay  int            `json:"debug_delay,omitempty"` // Délai d'attente pour la simulation de la concurrence
}

// LogType représente le type de log à afficher utilisé par une "enum" contenant INFO, ERROR, DEBUG et LAMPORT.
type LogType string

const (
	INFO    LogType = "INFO"
	ERROR   LogType = "ERROR"
	DEBUG   LogType = "DEBUG"
	LAMPORT LogType = "LAMPORT"
)

// CommunicationType représente le type de communication utilisé par une "enum" contenant Request, Acknowledge, Release.
type CommunicationType string

const (
	Request     CommunicationType = "REQ"
	Acknowledge CommunicationType = "ACK"
	Release     CommunicationType = "REL"
)

// Communication représente une communication pour l'algorithme de Lamport optimisé entre deux serveurs.
type Communication struct {
	Type    CommunicationType `json:"type"`              // Type de communication
	From    int               `json:"from"`              // Numéro du serveur émetteur
	To      []int             `json:"to"`                // Numéro des serveurs récepteurs
	Stamp   int               `json:"stamp"`             // Estampille associée à la communication
	Payload map[int]Event     `json:"payload,omitempty"` // Payload éventuel de la communication
}

// Command est un type représentant une commande valide à envoyer par un client au serveur.
type Command struct {
	Name       string // Nom de la commande
	Auth       bool   // Indique si la commande nécessite des credentials
	MinArgs    int    // Nombre minimum d'arguments
	MinOptArgs int    // Nombre minimum d'arguments optionnels
}

// User est un type représentant un utilisateur pouvant être un organisateur de manifestations ou un bénévole s'inscrivant à des jobs.
type User struct {
	Username string `json:"username"` // Nom d'utilisateur
	Password string `json:"password"` // Mot de passe
}

// Job est un type représentant un job lié à une manifestation.
type Job struct {
	Name         string `json:"name"`          // Nom du job
	NbVolunteers int    `json:"nb_volunteers"` // Nombre de bénévoles requis
	VolunteerIds []int  `json:"volunteer_ids"` // Liste des ids des bénévoles inscrits
}

// Event est un type représentant une manifestation.
type Event struct {
	Name      string      `json:"name"`       // Nom de la manifestation
	Closed    bool        `json:"closed"`     // Indique si la manifestation est fermée
	CreatorId int         `json:"creator_id"` // Id de l'organisateur
	Jobs      map[int]Job `json:"jobs"`       // Liste des jobs de la manifestation
}

// Entities est un type représentant les entités du serveur.
// Il contient les utilisateurs et les manifestations stockés dans des maps.
type Entities struct {
	Users  map[int]User  `json:"users"`  // Liste des utilisateurs
	Events map[int]Event `json:"events"` // Liste des manifestations
}
