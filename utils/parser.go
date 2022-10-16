package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/Lazzzer/labo1-sdr/utils/types"
)

// Aliases
type Config = types.Config
type Entities = types.Entities
type Event = types.Event
type Job = types.Job
type User = types.User

func GetEntities(content string) (map[int]User, map[int]Event, map[int]Job) {
	entities := parse[Entities](content)
	return entities.Users, entities.Events, entities.Jobs
}

func GetConfig(path string) Config {
	return *parse[Config](path)
}

func parse[T Config | Entities](content string) *T {
	var object T

	err := json.Unmarshal([]byte(content), &object)

	if err != nil {
		fmt.Println(err)
		panic("Error: Could not parse object")
	}

	return &object
}
