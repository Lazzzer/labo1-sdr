// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 1 SDR

// Package main est le point d'entrée du programme permettant de démarrer le serveur ou un client.
// Il gère aussi les flags du serveur pour le lancer en mode "debug" ou em mode "silent".
package main

import (
	_ "embed"
	"flag"
	"github.com/Lazzzer/labo1-sdr/server"
	"github.com/Lazzzer/labo1-sdr/utils"
	"log"
	"os"
	"strconv"
)

//go:embed config.json
var configJson string

// main est la méthode d'entrée du programme
func main() {

	if len(os.Args) != 2 {
		log.Fatal("Invalid number of arguments, usage: go run ./mainServer <number of the server>")
	}

	config := utils.GetConfig(configJson)

	if number, err := strconv.Atoi(os.Args[1]); err == nil && checkNumberOfServer(number, len(config.Ports)) {
		config.Port = config.Ports[number]
	} else {
		log.Fatal("Invalid server number")
	}

	debug := flag.Bool("debug", false, "Boolean: Run server in debug mode. Default is false")
	silent := flag.Bool("silent", false, "Boolean: Run server in silent mode. Default is false")

	flag.Parse()

	if *debug {
		config.Debug = true
	}

	if *silent {
		config.Silent = true
	}

	serv := server.Server{Config: config}
	serv.Run()
}

func checkNumberOfServer(maxValue int, value int) bool {
	return value >= 0 && value < maxValue
}
