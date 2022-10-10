package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Lazzzer/labo1-sdr/utils/types"
)

// Aliases
type Config = types.Config
type Entities = types.Entities
type Event = types.Event
type Job = types.Job
type User = types.User

func GetEntities(path string) ([]User, []Event, []Job) {
	entities := parse[Entities](path)
	return entities.Users, entities.Events, entities.Jobs
}

func GetConfig(path string) Config {
	return *parse[Config](path)
}

func convertJsonFileToByteArray(path string) []byte {
	jsonFile, err := os.Open(path)

	if err != nil {
		fmt.Println(err)
		panic("Error: Could not open file")
	}

	defer jsonFile.Close()

	byteValue, errIO := io.ReadAll(jsonFile)

	if errIO != nil {
		fmt.Println(errIO)
		// Not fully sure how defer works if there is a panic here.
		panic("Error: Could not read file")
	}

	return byteValue
}

func parse[T Config | Entities](path string) *T {
	byteValue := convertJsonFileToByteArray(path)

	var object T
	err := json.Unmarshal(byteValue, &object)

	if err != nil {
		fmt.Println(err)
		panic("Error: Could not parse object")
	}

	return &object
}
