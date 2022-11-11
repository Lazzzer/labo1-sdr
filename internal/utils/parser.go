// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 1 SDR

// Package utils contient des fichiers utilitaires pour le projet.
// Il contient notamment des variables globales de string pour les couleurs, les variables représentant les commandes
// et les messages formatés pour les réponses du serveur ou du client.
// Finalement, il contient un parser qui peut lire un fichier json et retourner soit la configuration soit les entités.
package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/Lazzzer/labo1-sdr/internal/utils/types"
)

// GetEntities parse une string pour retourner un tuple contenant les entités créées.
func GetEntities(content string) (map[int]types.User, map[int]types.Event) {
	entities := parse[types.Entities](content)
	return entities.Users, entities.Events
}

// GetConfig parse une string et retourne une configuration.
func GetConfig[T types.Config | types.ServerConfig](path string) T {
	return *parse[T](path)
}

// parse est une fonction générique limitée aux types Config et Entities et permet de retourner le pointeur de l'objet parsé
func parse[T types.Config | types.ServerConfig | types.Entities](content string) *T {
	var object T

	err := json.Unmarshal([]byte(content), &object)

	if err != nil {
		fmt.Println(err)
		panic("Error: Could not parse object")
	}

	return &object
}
