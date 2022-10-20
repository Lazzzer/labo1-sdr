// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 1 SDR
// source: https://twin.sh/articles/35/how-to-add-colors-to-your-console-terminal-output-in-go

package utils

var RESET = "\033[0m"         // Variable pour réinitialiser la couleur du texte
var RED = "\033[31m"          // Variable pour colorer le texte en rouge
var GREEN = "\033[32m"        // Variable pour colorer le texte en vert
var YELLOW = "\033[33m"       // Variable pour colorer le texte en jaune
var ORANGE = "\033[38;5;208m" // Variable pour colorer le texte en orange
var CYAN = "\033[36m"         // Variable pour colorer le texte en cyan
var BOLD = "\033[1m"          // Variable pour changer le texte en gras
