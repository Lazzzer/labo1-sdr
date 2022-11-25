// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 2 SDR

// Package main est le point d'entrée du programme permettant de démarrer le serveur.
// Il gère aussi les flags du serveur pour le lancer en mode "debug" ou em mode "silent".
package main

import (
	_ "embed"
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/Lazzzer/labo1-sdr/internal/server"
	"github.com/Lazzzer/labo1-sdr/internal/utils"
	"github.com/Lazzzer/labo1-sdr/internal/utils/types"
)

//go:embed config.json
var config string

// main est la méthode d'entrée du programme
func main() {

	debug := flag.Bool("debug", false, "Boolean: Run server in debug mode. Default is false")
	silent := flag.Bool("silent", false, "Boolean: Run server in silent mode. Default is false")

	flag.Parse()

	if flag.Arg(0) == "" {
		log.Fatal("Invalid argument, usage: -debug -silent <server number>")
	}

	number, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		log.Fatal("Invalid argument, usage: -debug -silent <server number>")
	}

	config := utils.GetConfig[types.ServerConfig](config)

	if address, ok := config.Servers[number]; ok {
		config.Address = address
	} else {
		log.Fatal("Invalid server number")
	}

	if *debug {
		config.Debug = true
	}

	if *silent {
		config.Silent = true
	}

	serv := server.Server{Number: number, Port: strings.Split(config.Address, ":")[1], ClientPort: config.ClientPorts[number], Config: config}
	serv.Run()
}
