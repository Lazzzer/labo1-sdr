package main

import (
	_ "embed"
	"flag"
	"fmt"

	"github.com/Lazzzer/labo1-sdr/client"
	"github.com/Lazzzer/labo1-sdr/server"
	"github.com/Lazzzer/labo1-sdr/utils"
)

//go:embed config.json
var configJson string

func main() {

	serverMode := flag.Bool("server", false, "Boolean: Run program in server mode. Default is client mode")
	debug := flag.Bool("debug", false, "Boolean: Run server in debug mode. Default is false")
	silent := flag.Bool("silent", false, "Boolean: Run server in silent mode. Default is false")

	flag.Parse()

	config := utils.GetConfig(configJson)

	if *serverMode {

		if *debug {
			config.Debug = true
		}

		if *silent {
			config.Silent = true
		}

		server := server.Server{Config: config}
		server.Run()
	} else {

		if *debug || *silent {
			fmt.Println("Error: Debug and silent mode are only available in server mode, use --help or -h for more information")
			return
		}

		client := client.Client{Config: config}
		client.Run()
	}
}
