// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 1 SDR

// Package main est le point d'entrée du programme permettant de démarrer le serveur ou un client.
// Il gère aussi les flags du serveur pour le lancer en mode "debug" ou em mode "silent".
package main

import (
	_ "embed"
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/Lazzzer/labo1-sdr/internal/client"
	"github.com/Lazzzer/labo1-sdr/internal/utils"
	"github.com/Lazzzer/labo1-sdr/internal/utils/types"
)

//go:embed config.json
var config string

func main() {
	number := flag.Int("number", -1, "Integer: Number of the server to connect to, Default is -1")
	flag.Parse()

	if flag.Arg(0) == "" {
		log.Fatal("Invalid argument, usage: -number=1 <client name>")
	}

	config := utils.GetConfig[types.Config](config)

	if *number == -1 {
		rand.Seed(time.Now().UnixNano())
		randomPos := rand.Intn(len(config.Servers) + 1)
		config.Address = config.Servers[randomPos]
	} else if address, ok := config.Servers[*number]; ok {
		config.Address = address
	} else {
		log.Fatal("Invalid server number")
	}

	cl := client.Client{Name: flag.Arg(0), Config: config}
	cl.Run()
}
