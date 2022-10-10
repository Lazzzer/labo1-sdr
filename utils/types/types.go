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
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Job struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	CreatorId    int    `json:"creator_id"`
	NbVolunteers int    `json:"nb_volunteers"`
	VolunteerIds []int  `json:"volunteer_ids"`
}

type Event struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	CreatorId int    `json:"creator_id"`
	JobIds    []int  `json:"job_ids"`
}

type Entities struct {
	Users  []User  `json:"users"`
	Events []Event `json:"events"`
	Jobs   []Job   `json:"jobs"`
}
