// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 2 SDR

// Package client propose un client TCP qui se connecte à un serveur gestionnaire de manifestations.
//
// L'URL et le port du serveur sont définis dans le fichier config.json injecté dans l'application au build.
// Le client peut spécifier le numéro du serveur auquel il souhaite se connecter et s'il ne le fait pas, il se connecte à
// au hasard à un serveur présent dans la liste.
// Le client est capable d'envoyer des commandes au serveur et d'afficher ses réponses.
// Les commandes protégées par des credentials activent un prompt pour y passer ses identifiants.
// Les commandes qui n'existent pas ou contenant des typos (par exemple: "shutdownServer" ou "helpp") ne sont même pas envoyées au serveur.
// Un CTRL+C signale quand même au serveur que le client se déconnecte et le client se termine "gracefully".
package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Lazzzer/labo1-sdr/internal/utils"
	"github.com/Lazzzer/labo1-sdr/internal/utils/types"
	"golang.org/x/term"
)

// Client est une struct représentant un client TCP.
type Client struct {
	Name   string       // Nom du client
	Config types.Config // Configuration du client
}

// Run lance le client et se connecte à un serveur.
//
// Les réponses du serveur et le CTRL+C sont gérés par des goroutines.
// Lorsqu'un utilisateur souhaite quitter l'application, le client envoie quoiqu'il arrive un message au serveur pour fermer la connexion.
func (c *Client) Run() {

	intChan := make(chan os.Signal, 1) // Catch du CTRL+C
	signal.Notify(intChan, syscall.SIGINT)

	conn, err := net.Dial("tcp", c.Config.Address)

	if err != nil {
		log.Fatal("❌ " + utils.RED + "Could not connect to the server." + utils.RESET)
	} else {
		fmt.Println(utils.MESSAGE.Title)
	}

	_, err = conn.Write([]byte(c.Name + "\n"))
	if err != nil {
		log.Fatal("❌ " + utils.RED + "Could not send name to the server." + utils.RESET)
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)

	go func() {
		<-intChan
		_, err = conn.Write([]byte("quit\n"))
		if err != nil {
			log.Println(err)
		}
		fmt.Println(utils.MESSAGE.Goodbye)
		os.Exit(0)
	}()

	go func() {
		_, errFrom := io.Copy(os.Stdout, conn) // Lecture des réponses du serveur
		if errFrom != nil {
			os.Exit(0)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		processedInput, err := c.processInput(input)

		if err != nil {
			if err.Error() == "invalid input" {
				fmt.Print(utils.MESSAGE.Error.InvalidCommand)
			}
			continue
		}

		_, err = io.Copy(conn, strings.NewReader(processedInput+"\n")) // Passage de l'input traité au serveur
		if err != nil {
			log.Println(err)
		}

		if processedInput == utils.QUIT.Name {
			fmt.Println(utils.MESSAGE.Goodbye)
			break
		}
	}
}

// askCredentials crée un prompt et attend l'input de l'utilisateur pour son username et son password.
// L'insertion du password est en mode sans echo.
func (c *Client) askCredentials() (string, error) {
	fmt.Println(utils.MESSAGE.LoginStart)
	defer fmt.Println(utils.MESSAGE.LoginEnd)

	fmt.Print(utils.BOLD + "Enter Username: " + utils.RESET)
	username, errUsername := bufio.NewReader(os.Stdin).ReadString('\n')
	usernameArr := strings.Fields(username)

	if errUsername != nil || len(usernameArr) != 1 {
		return "", fmt.Errorf("invalid username")
	}
	username = usernameArr[0]

	fmt.Print(utils.BOLD + "Enter Password: " + utils.RESET)
	bytePassword, errPassword := term.ReadPassword(int(syscall.Stdin))

	if errPassword != nil {
		return "", errPassword
	}

	return username + " " + string(bytePassword), nil
}

// processInput traite l'input de l'utilisateur et vérifie si l'input peut être mappé à une commande.
// La méthode vérifie aussi si une authentification est nécessaire et s'il y a une entrée vide.
func (c *Client) processInput(input string) (string, error) {
	args := strings.Fields(input)

	if len(args) == 0 {
		return "", fmt.Errorf("empty input")
	}

	processedInput := strings.Join(args, " ")

	for _, command := range utils.COMMANDS {
		if args[0] == command.Name {
			if command.Auth {
				credentials, err := c.askCredentials()
				if err != nil {
					return "", fmt.Errorf("invalid input")
				}
				processedInput += " " + credentials
			}
			return processedInput, nil
		}
	}
	return "", fmt.Errorf("invalid input")
}
