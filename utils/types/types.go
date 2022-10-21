// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 1 SDR

// Package types propose différents types utilisés par l'application pour parser les fichiers de configuration et les entités.
package types

// Config est un type représentant la configuration d'un client ou d'un serveur.
// Le client ne se soucie que des champs Host et Port que le serveur utilise.
type Config struct {
	Host       string `json:"host,omitempty"`        // Adresse du serveur
	Port       int    `json:"port,omitempty"`        // Port du serveur
	Ports      []int  `json:"ports"`                 // Ports des serveurs disponibles
	Debug      bool   `json:"debug,omitempty"`       // Activation du mode debug pour vérifier la concurrence
	Silent     bool   `json:"silent,omitempty"`      // Activation du mode silencieux pour ne pas afficher les logs
	DebugDelay int    `json:"debug_delay,omitempty"` // Délai d'attente pour la simulation de la concurrence
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
