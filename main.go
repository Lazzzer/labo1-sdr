package main

import (
	"flag"
	"fmt"

	"github.com/Lazzzer/labo1-sdr/client"
	"github.com/Lazzzer/labo1-sdr/server"
	"github.com/Lazzzer/labo1-sdr/utils"
)

func main() {

	serverMode := flag.Bool("server", false, "Boolean: Run program in server mode. Default is client mode")
	debug := flag.Bool("debug", false, "Boolean: Run server in debug mode. Default is false")

	flag.Parse()

	if *serverMode {
		config := utils.GetConfig("server/config.json")

		if *debug {
			config.Debug = true
		}

		server := server.Server{Config: config}
		server.Run()
	} else {

		if *debug {
			fmt.Println("Error: Debug mode is only available in server mode, use --help or -h for more information")
			return
		}

		config := utils.GetConfig("client/config.json")
		client := client.Client{Config: config}
		client.Run()
	}
}
