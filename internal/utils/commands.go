// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 2 SDR

package utils

import "github.com/Lazzzer/labo1-sdr/internal/utils/types"

var HELP = types.Command{Name: "help", Auth: false, MinArgs: 0, MinOptArgs: -1}        // Propriétés de la commande "help"
var CREATE = types.Command{Name: "create", Auth: true, MinArgs: 5, MinOptArgs: 2}      // Propriétés de la commande "create"
var CLOSE = types.Command{Name: "close", Auth: true, MinArgs: 3, MinOptArgs: -1}       // Propriétés de la commande "close"
var REGISTER = types.Command{Name: "register", Auth: true, MinArgs: 4, MinOptArgs: -1} // Propriétés de la commande "register"
var SHOW = types.Command{Name: "show", Auth: false, MinArgs: 0, MinOptArgs: 1}         // Propriétés de la commande "show"
var JOBS = types.Command{Name: "jobs", Auth: false, MinArgs: 1, MinOptArgs: -1}        // Propriétés de la commande "jobs"
var QUIT = types.Command{Name: "quit", Auth: false, MinArgs: 0, MinOptArgs: -1}        // Propriétés de la commande "quit"

var COMMANDS = [...]types.Command{
	HELP,
	CREATE,
	CLOSE,
	REGISTER,
	SHOW,
	JOBS,
	QUIT,
}
