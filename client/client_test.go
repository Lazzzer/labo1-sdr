package client

import (
	"net"
	"strconv"
	"testing"

	"github.com/Lazzzer/labo1-sdr/server"
	"github.com/Lazzzer/labo1-sdr/utils"
)

var testingConfig = utils.Config{Host: "localhost", Port: 8081}

type TestInput struct {
	Description string
	Input       string
	Expected    string
}

type TestClient struct {
	Config utils.Config
}

func init() {
	server := server.Server{Config: utils.Config{Port: testingConfig.Port, Debug: true}}
	go server.Run()
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

func Test_Help_Command(t *testing.T) {
	testClient := TestClient{Config: testingConfig}

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
			Description: "Send help command and receive help message",
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

func Test_Close_Command(t *testing.T) {
	testClient := TestClient{Config: testingConfig}

	tests := []TestInput{
		{
			Description: "Send invalid help command and receive error message",
			Input:       "closee 1 johndoe 1234\n",
			Expected:    "Error: Invalid command. Type 'help' for a list of commands.\n",
		},
		{
			Description: "Send close command and receive confirmation message",
			Input:       "close 1 johndoe 1234\n",
			Expected:    "Event with id 1 is closed.\n",
		},
		{
			Description: "Send close command on closed event and receive error message",
			Input:       "close 1 johndoe 1234\n",
			Expected:    "Error: Event is already closed.\n",
		},
	}
	testClient.Run(tests, t)
}
