package types

type Config struct {
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port"`
	Debug  bool   `json:"debug,omitempty"`
	Silent bool   `json:"silent,omitempty"`
}

type Command struct {
	Name       string
	Auth       bool
	MinArgs    int
	MinOptArgs int
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Job struct {
	Name         string `json:"name"`
	CreatorId    int    `json:"creator_id"`
	EventId      int    `json:"event_id"`
	NbVolunteers int    `json:"nb_volunteers"`
	VolunteerIds []int  `json:"volunteer_ids"`
}

type Event struct {
	Name      string `json:"name"`
	Closed    bool   `json:"closed"`
	CreatorId int    `json:"creator_id"`
	JobIds    []int  `json:"job_ids"`
}

type Entities struct {
	Users  map[int]User  `json:"users"`
	Events map[int]Event `json:"events"`
	Jobs   map[int]Job   `json:"jobs"`
}
