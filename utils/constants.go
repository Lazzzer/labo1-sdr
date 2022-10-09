package utils

import "github.com/Lazzzer/labo1-sdr/utils/types"

var COMMANDS = [...]types.Command{
	{Name: "help", Auth: false, MinArgs: 0},
	{Name: "create", Auth: true, MinArgs: 4},
	{Name: "close", Auth: true, MinArgs: 2},
	{Name: "register", Auth: true, MinArgs: 3},
	{Name: "showAll", Auth: false, MinArgs: 0},
	{Name: "showJobs", Auth: false, MinArgs: 1},
	{Name: "jobRepartition", Auth: false, MinArgs: 1},
	{Name: "quit", Auth: false, MinArgs: 0},
}
