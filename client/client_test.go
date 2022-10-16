package client

import (
	"net"
	"strconv"
	"testing"

	"github.com/Lazzzer/labo1-sdr/utils"
)

type TestInput struct {
	Description string
	Input       string
	Expected    string
}

type TestClient struct {
	Config utils.Config
}

func (tc *TestClient) Run(tests []TestInput, t *testing.T) {
	conn, err := net.Dial("tcp", tc.Config.Host+":"+strconv.Itoa(tc.Config.Port))

	if err != nil {
		t.Error("Error: could not connect to server")
		return
	}

	defer conn.Close()

	for _, test := range tests {
		if _, err := conn.Write([]byte(test.Input)); err != nil {
			t.Error("Error: could not write to server")
		}

		out := make([]byte, 1024)
		if _, err := conn.Read(out); err == nil {
			if string(out[:len(test.Expected)]) != test.Expected {
				t.Error("\nError for test: ", test.Description, "\n\nResponse did match expected output:\n>> Expected\n", test.Expected, "\n>> Got\n", string(out))
			}
		} else {
			t.Error("Error: could not read from connection")
		}
	}

	if _, err := conn.Write([]byte("quit\n")); err != nil {
		t.Error("Error: could not quit the server properly")
	}
}

func TestClient_Help_Command(t *testing.T) {
	testClient := TestClient{Config: utils.Config{Host: "localhost", Port: 8080}}

	var help = "---------------------------------------------------------\n"
	help += "# Description of the command:\nHow to use the command\n \n"
	help += "# Display all commands:\nhelp\n \n"
	help += "# Create an event (will ask your password):\ncreate <eventName> <username> <job1> <nbVolunteer1> [<job2> <nbVolunteer2> ...]\n \n"
	help += "# Close an event (will ask your password):\nclose <idEvent> <username>\n \n"
	help += "# Register to an event (will ask your password):\nregister <idEvent> <idJob> <username>\n \n"
	help += "# Show all events:\nshowAll\n \n"
	help += "# Show all jobs from an event:\nshowJobs <idEvent>\n \n"
	help += "# Show all volunteers from an event:\njobRepartition <idEvent>\n \n"
	help += "# Quit the program:\nquit\n"
	help += "---------------------------------------------------------\n"

	tests := []TestInput{
		{
			Description: "Send valid help command and receive help message",
			Input:       "help\n",
			Expected:    help,
		},
		{
			Description: "Send invalid help command and receive error message",
			Input:       "helpp\n",
			Expected:    "Error: Invalid command. Type 'help' for a list of commands.\n",
		},
		// {
		// 	Description: "Failling test",
		// 	Input:       "helpp\n",
		// 	Expected:    "Blabla\n",
		// },
	}

	testClient.Run(tests, t)
}
