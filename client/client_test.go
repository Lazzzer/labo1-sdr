package client

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/Lazzzer/labo1-sdr/server"
	"github.com/Lazzzer/labo1-sdr/utils"
)

var testingConfig = utils.Config{Host: "localhost", Port: 8081}
var testingDebugConfig = utils.Config{Host: "localhost", Port: 8082}

type TestInput struct {
	Description string
	Input       string
	Expected    string
}

type TestClient struct {
	Config utils.Config
}

func init() {
	serv := server.Server{Config: utils.Config{Port: 8081, Debug: false, Silent: true}}
	servDebug := server.Server{Config: utils.Config{Port: 8082, Debug: true, Silent: true}}

	go serv.Run()
	go servDebug.Run()
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

func run(tc *TestClient, wg *sync.WaitGroup, tests []TestInput, t *testing.T) {
	defer wg.Done()
	fmt.Println("Test client is running")
	tc.Run(tests, t)
}

func runConcurrent(nbClients int, tests []TestInput, t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(nbClients)
	for i := 0; i < nbClients; i++ {
		testClient := TestClient{Config: testingDebugConfig}
		go run(&testClient, &wg, tests, t)
	}
	wg.Wait()
	fmt.Println("Test clients are done")
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

func Test_Create_Command(t *testing.T) {
	// TODO
}

func Test_Close_Command(t *testing.T) {
	testClient := TestClient{Config: testingConfig}

	tests := []TestInput{
		{
			Description: "Send close command and receive confirmation message",
			Input:       "close 3 jane root\n",
			Expected:    "Event with id 3 is closed.\n",
		},
		{
			Description: "Send close command with bad id and receive error message",
			Input:       "close bad lazar root\n",
			Expected:    "Error: event id must be integer.\n",
		},
		{
			Description: "Send close command with bad credentials and receive receive error message",
			Input:       "close 3 jane rooot\n",
			Expected:    "Error: Access denied.\n",
		},
		{
			Description: "Send close command on closed event and receive error message",
			Input:       "close 1 claude root\n",
			Expected:    "Error: Event is already closed.\n",
		},
		{
			Description: "Send close command on event not owned by the user and receive error message",
			Input:       "close 2 claude root\n",
			Expected:    "Error: Only the creator of the event can close it.\n",
		},
		{
			Description: "Send invalid close command and receive error message",
			Input:       "closee 1 claude root\n",
			Expected:    "Error: Invalid command. Type 'help' for a list of commands.\n",
		},
	}
	testClient.Run(tests, t)
}

func Test_Register_Command(t *testing.T) {
	testClient := TestClient{Config: testingConfig}

	tests := []TestInput{
		{
			Description: "Send register command and receive confirmation message",
			Input:       "register 2 4 lazar root\n",
			Expected:    "User registered to job with id 4 in event Baleinev 2023.\n",
		},
		{
			Description: "Send register command for a user in another job of the same event and receive confirmation message",
			Input:       "register 2 6 lazar root\n",
			Expected:    "User registered to job with id 6 in event Baleinev 2023.\n",
		},
		{
			Description: "Send register command for a user who left a job for another and came back to the job in the same event and receive confirmation message",
			Input:       "register 2 4 lazar root\n",
			Expected:    "User registered to job with id 4 in event Baleinev 2023.\n",
		},
		{
			Description: "Send register command with bad ids and receive error message",
			Input:       "register bad id lazar root\n",
			Expected:    "Error: Ids must be integers.\n",
		},
		{
			Description: "Send register command for inexistant event and receive error message",
			Input:       "register 1000 1 valentin root\n",
			Expected:    "Error: Event not found by this id.\n",
		},
		{
			Description: "Send register command for inexistant job for the given event and receive error message",
			Input:       "register 2 1000 valentin root\n",
			Expected:    "Error: Job not found with given id.\n",
		},
		{
			Description: "Send register command with bad credentials and receive receive error message",
			Input:       "register 2 4 jane rooot\n",
			Expected:    "Error: Access denied.\n",
		},
		{
			Description: "Send register command with user already in job and receive error message",
			Input:       "register 2 4 valentin root\n",
			Expected:    "Error: User is already registered in this job.\n",
		},
		{
			Description: "Send register command for job already full and receive error message",
			Input:       "register 2 5 claude root\n",
			Expected:    "Error: Job is already full.\n",
		},
		{
			Description: "Send register command as creator of event and receive error message",
			Input:       "register 2 4 john root\n",
			Expected:    "Error: Creator of the event cannot register for a job.\n",
		},
		{
			Description: "Send register command on closed event and receive error message",
			Input:       "register 1 1 lazar root\n",
			Expected:    "Error: Event is closed.\n",
		},
		{
			Description: "Send invalid register command and receive error message",
			Input:       "registeer bad input\n",
			Expected:    "Error: Invalid command. Type 'help' for a list of commands.\n",
		},
	}
	testClient.Run(tests, t)
}

func Test_Show_Command(t *testing.T) {
	// TODO
}

func Test_Commands_Concurrently(t *testing.T) {
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

	var showAll = "Events:\n"
	showAll += "Montreux Jazz 2022\n"
	showAll += "Baleinev 2023\n"
	showAll += "Balélec 2023\n"

	tests := []TestInput{
		{
			Description: "Send help command and receive help message",
			Input:       "help\n",
			Expected:    help,
		},
		{
			Description: "Send show command and receive message",
			Input:       "showAll\n",
			Expected:    showAll,
		},
	}

	runConcurrent(2, tests, t)
}
