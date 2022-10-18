package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/Lazzzer/labo1-sdr/utils/types"
)

func GetEntities(content string) (map[int]types.User, map[int]types.Event) {
	entities := parse[types.Entities](content)
	return entities.Users, entities.Events
}

func GetConfig(path string) types.Config {
	return *parse[types.Config](path)
}

func parse[T types.Config | types.Entities](content string) *T {
	var object T

	err := json.Unmarshal([]byte(content), &object)

	if err != nil {
		fmt.Println(err)
		panic("Error: Could not parse object")
	}

	return &object
}
