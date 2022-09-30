package main

import (
	"fmt"

	"github.com/Lazzzer/labo1-sdr/utils"
)

func main() {
	users, manifestations := utils.GetEntities("../utils/config.json")
	fmt.Println(users)
	fmt.Println(manifestations)
}
