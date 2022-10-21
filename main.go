// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 1 SDR

// Package main est le point d'entrée du programme permettant de démarrer le serveur ou un client.
// Il gère aussi les flags du serveur pour le lancer en mode "debug" ou em mode "silent".
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"math/rand"

	"github.com/Lazzzer/labo1-sdr/client"
	"github.com/Lazzzer/labo1-sdr/server"
	"github.com/Lazzzer/labo1-sdr/utils"
)

//go:embed config.json
var configJson string

// main est la méthode d'entrée du programme
func main() {

	serverMode := flag.Bool("server", false, "Boolean: Run program in server mode. Default is client mode")
	debug := flag.Bool("debug", false, "Boolean: Run server in debug mode. Default is false")
	silent := flag.Bool("silent", false, "Boolean: Run server in silent mode. Default is false")
	number := flag.Int("number", -1, "Integer: Number of the server to run or to connect to, Default is 0")
	name := flag.String("name", "", "String: Name of the client to run, Default is empty string")

	flag.Parse()

	config := utils.GetConfig(configJson)

	if *serverMode {

		if *name != "" {
			log.Fatal("Server doesn't need a name")
		}

		if *number < len(config.Ports) && *number >= 0 {
			config.Port = config.Ports[*number]
		} else {
			log.Fatal("Invalid server number")
		}

		if *debug {
			config.Debug = true
		}

		if *silent {
			config.Silent = true
		}

		serv := server.Server{Config: config}
		serv.Run()
	} else {

		if *debug || *silent {
			fmt.Println("Error: Debug and silent mode are only available in server mode, use --help or -h for more information")
			return
		}

		if *number == -1 {
			if *number < len(config.Ports) && *number >= 0 {
				config.Port = config.Ports[*number]
			} else {
				log.Fatal("Invalid server number")
			}
		} else {
			randomPos := rand.Intn(len(config.Ports))
			config.Port = config.Ports[randomPos]
		}

		if *name != "" {
			cl := client.Client{Name: *name, Config: config}
			cl.Run()
		} else {
			log.Fatal("Client needs a name")
		}

	}
}
