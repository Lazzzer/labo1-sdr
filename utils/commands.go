package utils

import "github.com/Lazzzer/labo1-sdr/utils/types"

var HELP = types.Command{Name: "help", Auth: false, MinArgs: 0, MinOptArgs: -1}
var CREATE = types.Command{Name: "create", Auth: true, MinArgs: 5, MinOptArgs: 2}
var CLOSE = types.Command{Name: "close", Auth: true, MinArgs: 3, MinOptArgs: -1}
var REGISTER = types.Command{Name: "register", Auth: true, MinArgs: 4, MinOptArgs: -1}
var SHOW = types.Command{Name: "show", Auth: false, MinArgs: 0, MinOptArgs: 1}
var JOBS = types.Command{Name: "jobs", Auth: false, MinArgs: 1, MinOptArgs: -1}
var QUIT = types.Command{Name: "quit", Auth: false, MinArgs: 0, MinOptArgs: -1}

var COMMANDS = [...]types.Command{
	HELP,
	CREATE,
	CLOSE,
	REGISTER,
	SHOW,
	JOBS,
	QUIT,
}
