package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Lazzzer/labo1-sdr/utils/types"
)

func convertJsonToByteArray(path string) []byte {
	jsonFile, err := os.Open(path)

	if err != nil {
		fmt.Println(err)
		panic("Error: Could not open config file")
	}

	defer jsonFile.Close()

	byteValue, errIO := io.ReadAll(jsonFile)

	if errIO != nil {
		fmt.Println(errIO)
		// Not fully sure how defer works if there is a panic here.
		panic("Error: Could not read config file")
	}

	return byteValue
}

func parseConfig(path string) *types.Config {
	byteValue := convertJsonToByteArray(path)

	var config types.Config

	err := json.Unmarshal(byteValue, &config)

	if err != nil {
		fmt.Println(err)
		panic("Error: Could not parse config")
	}

	return &config
}

func GetEntities(path string) ([]types.User, []types.Manifestation) {
	config := parseConfig(path)
	return config.Users, config.Manifestations
}
