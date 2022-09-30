package types

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Job struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	NbVolunteers int    `json:"nb_volunteers"`
}

type Manifestation struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	CreatorName     string `json:"creator_name"`
	CreatorPassword string `json:"creator_password"`
	Jobs            []Job  `json:"jobs"`
}

type Config struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port"`
}

type Entities struct {
	Users          []User          `json:"users"`
	Manifestations []Manifestation `json:"manifestations"`
}
