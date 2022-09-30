package main

import (
	"fmt"

	"github.com/Lazzzer/labo1-sdr/utils"
)

func main() {
	config := utils.GetConfig("config.json")
	fmt.Println(config)
}
