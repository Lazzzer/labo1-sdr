// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 1 SDR

// Package main est le point d'entrée du programme permettant de démarrer le serveur ou un client.
// Il gère aussi les flags du serveur pour le lancer en mode "debug" ou em mode "silent".
package main

import (
	_ "embed"
	"flag"
	"github.com/Lazzzer/labo1-sdr/client"
	"github.com/Lazzzer/labo1-sdr/utils"
	"log"
	"math/rand"
	"os"
)

//go:embed config.json
var configJsonClient string

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Invalid number of arguments, usage: go run ./mainClient.go <name of the client>", len(os.Args))
	}
	if os.Args[1] == "" {
		log.Fatal("Invalid name")
	}

	config := utils.GetConfig(configJsonClient)

	number := flag.Int("number", -1, "Integer: Number of the server to run or to connect to, Default is -1")
	flag.Parse()

	if *number == -1 {
		randomPos := rand.Intn(len(config.Ports))
		config.Port = config.Ports[randomPos]
	} else if *number >= 0 && *number < len(config.Ports) {
		config.Port = config.Ports[*number]
	} else {
		log.Fatal("Invalid server number")
	}

	cl := client.Client{Name: os.Args[1], Config: config}
	cl.Run()
}
