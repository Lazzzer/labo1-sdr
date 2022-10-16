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
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Job struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	CreatorId    int    `json:"creator_id"`
	EventId      int    `json:"event_id"`
	NbVolunteers int    `json:"nb_volunteers"`
	VolunteerIds []int  `json:"volunteer_ids"`
}

type Event struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Closed    bool   `json:"closed"`
	CreatorId int    `json:"creator_id"`
	JobIds    []int  `json:"job_ids"`
}

type Entities struct {
	Users  []User  `json:"users"`
	Events []Event `json:"events"`
	Jobs   []Job   `json:"jobs"`
}
