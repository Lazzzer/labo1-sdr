package types

type Config struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port"`
}

type Command struct {
	Name    string `json:"name"`
	Auth    bool   `json:"auth"`
	MinArgs int    `json:"minArgs"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Job struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	NbVolunteers int    `json:"nb_volunteers"`
	Volunteers   []User `json:"volunteers"`
}

type Event struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	CreatorName     string `json:"creator_name"`
	CreatorPassword string `json:"creator_password"`
	Jobs            []Job  `json:"jobs"`
}

type Entities struct {
	Users  []User  `json:"users"`
	Events []Event `json:"events"`
}
