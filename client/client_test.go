package client

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/Lazzzer/labo1-sdr/server"
	"github.com/Lazzzer/labo1-sdr/utils"
	"github.com/Lazzzer/labo1-sdr/utils/types"
)

var testingConfig = types.Config{Host: "localhost", Port: 8081}
var testingDebugConfig = types.Config{Host: "localhost", Port: 8082}

type TestInput struct {
	Description string
	Input       string
	Expected    string
}

type TestClient struct {
	Config types.Config
}

func init() {
	serv := server.Server{Config: types.Config{Port: 8081, Debug: false, Silent: true}}
	servDebug := server.Server{Config: types.Config{Port: 8082, Debug: true, Silent: true, DebugDelay: 5}}

	go serv.Run()
	go servDebug.Run()
}

func (tc *TestClient) Run(tests []TestInput, t *testing.T) {
	conn, err := net.Dial("tcp", tc.Config.Host+":"+strconv.Itoa(tc.Config.Port))

	if err != nil {
		t.Error(utils.RED + "FAIL: " + utils.RESET + "Error: could not connect to server")
		return
	}

	defer conn.Close()

	for _, test := range tests {
		if _, err := conn.Write([]byte(test.Input)); err != nil {
			t.Error(utils.RED + "FAIL: " + utils.RESET + "Error: could not write to server")
		}

		out := make([]byte, 2048)
		if _, err := conn.Read(out); err == nil {
			if string(out[:len(test.Expected)]) != test.Expected {
				t.Error("\n" + utils.RED + "FAIL: " + utils.RESET + test.Description + utils.GREEN + "\n\nExpected\n" + utils.RESET + test.Expected + utils.RED + "\nReceived\n" + utils.RESET + string(out))
			} else {
				fmt.Println(utils.GREEN + "PASS: " + utils.RESET + test.Description)
			}
		} else {
			t.Error(utils.RED + "FAIL: " + utils.RESET + "Error: could not read from connection")
		}

	}

	if _, err := conn.Write([]byte("quit\n")); err != nil {
		t.Error(utils.RED + "FAIL: " + utils.RESET + "Error: could not quit the server properly")
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

	tests := []TestInput{
		{
			Description: "Send help command and receive help message",
			Input:       "help\n",
			Expected:    utils.MESSAGE.Help,
		},
		{
			Description: "Send invalid help command and receive error message",
			Input:       "helpp\n",
			Expected:    utils.MESSAGE.Error.InvalidCommand,
		},
	}
	testClient.Run(tests, t)
}

func Test_Show_Command(t *testing.T) {
	testClient := TestClient{Config: testingConfig}

	var showAll = utils.RED + "Closed" + utils.RESET + "\t#1 " + utils.BOLD + utils.CYAN + "Montreux Jazz 2022" + utils.RESET + " / Creator: claude\n\n" +
		utils.GREEN + "Open" + utils.RESET + "\t#2 " + utils.BOLD + utils.CYAN + "Baleinev 2023" + utils.RESET + " / Creator: john\n\n" +
		utils.GREEN + "Open" + utils.RESET + "\t#3 " + utils.BOLD + utils.CYAN + "Balélec 2023" + utils.RESET + " / Creator: jane\n"

	var showFirstEvent = "\x1b[36m\n====================== 📅 EVENT 📅 ===========================\n\n\x1b[0m#1 \x1b[1m\x1b[36mMontreux Jazz 2022\x1b[0m\n\nCreator: claude\n\n🦺\x1b[1m Jobs\x1b[0m\n\n\x1b[32m(1/4)\x1b[0m\tJob #1: Montage\n\x1b[32m(2/10)\x1b[0m\tJob #2: Stands\n\x1b[31m(2/2)\x1b[0m\tJob #3: Sécurité\n\n\x1b[36m==============================================================\x1b[0m\n\n"

	tests := []TestInput{
		{
			Description: "Send show command and receive all events",
			Input:       "show\n",
			Expected:    utils.MESSAGE.WrapEvent(showAll),
		},
		{
			Description: "Send show command with id of first event and receive that event",
			Input:       "show 1\n",
			Expected:    showFirstEvent,
		},
		{
			Description: "Send show command with invalid nb of args and receive error message",
			Input:       "show 1 1 1\n",
			Expected:    utils.MESSAGE.Error.InvalidCommand,
		},
		{
			Description: "Send invalid show command and receive error message",
			Input:       "showw\n",
			Expected:    utils.MESSAGE.Error.InvalidCommand,
		},
	}
	testClient.Run(tests, t)
}

func Test_Create_Command(t *testing.T) {
	testClient := TestClient{Config: testingConfig}
	tests := []TestInput{
		{
			Description: "Send create command for event with one job and receive confirmation message",
			Input:       "create Test TestJob 1 lazar root\n",
			Expected:    utils.MESSAGE.WrapSuccess("Event #4 Test and 1 job(s) created\n"),
		},
		{
			Description: "Send create command for event with 3 jobs and receive confirmation message",
			Input:       "create Test TestJob 1 TestJob2 1 TestJob3 1 lazar root\n",
			Expected:    utils.MESSAGE.WrapSuccess("Event #5 Test and 3 job(s) created\n"),
		},
		{
			Description: "Send create command with invalid nb of args and receive error message",
			Input:       "create Test lazar root\n",
			Expected:    utils.MESSAGE.Error.InvalidNbArgs,
		},
		{
			Description: "Send create command with invalid nb of volunteers and receive error message",
			Input:       "create Test TestJob Invalid lazar root\n",
			Expected:    utils.MESSAGE.Error.NbVolunteersInteger,
		},
		{
			Description: "Send create command with invalid credentials and receive error message",
			Input:       "create Test TestJob 1 lazar rooooot\n",
			Expected:    utils.MESSAGE.Error.AccessDenied,
		},
		{
			Description: "Send invalid create command and receive error message",
			Input:       "createe\n",
			Expected:    utils.MESSAGE.Error.InvalidCommand,
		},
	}
	testClient.Run(tests, t)
}

func Test_Close_Command(t *testing.T) {
	testClient := TestClient{Config: testingConfig}

	tests := []TestInput{
		{
			Description: "Send close command and receive confirmation message",
			Input:       "close 3 jane root\n",
			Expected:    utils.MESSAGE.WrapSuccess("Event #3 is closed.\n"),
		},
		{
			Description: "Send close command with bad id and receive error message",
			Input:       "close bad lazar root\n",
			Expected:    utils.MESSAGE.Error.MustBeInteger,
		},
		{
			Description: "Send close command with bad credentials and receive receive error message",
			Input:       "close 3 jane rooot\n",
			Expected:    utils.MESSAGE.Error.AccessDenied,
		},
		{
			Description: "Send close command on closed event and receive error message",
			Input:       "close 1 claude root\n",
			Expected:    utils.MESSAGE.Error.AlreadyClosed,
		},
		{
			Description: "Send close command on event not owned by the user and receive error message",
			Input:       "close 2 claude root\n",
			Expected:    utils.MESSAGE.Error.NotCreator,
		},
		{
			Description: "Send invalid close command and receive error message",
			Input:       "closee 1 claude root\n",
			Expected:    utils.MESSAGE.Error.InvalidCommand,
		},
	}
	testClient.Run(tests, t)
}

func Test_Register_Command(t *testing.T) {
	testClient := TestClient{Config: testingConfig}

	tests := []TestInput{
		{
			Description: "Send register command and receive confirmation message",
			Input:       "register 2 1 lazar root\n",
			Expected:    utils.MESSAGE.WrapSuccess("User registered in job #1 for Event #2 Baleinev 2023.\n"),
		},
		{
			Description: "Send register command for a user in another job of the same event and receive confirmation message",
			Input:       "register 2 3 lazar root\n",
			Expected:    utils.MESSAGE.WrapSuccess("User registered in job #3 for Event #2 Baleinev 2023.\n"),
		},
		{
			Description: "Send register command for a user who left a job for another and came back to the job in the same event and receive confirmation message",
			Input:       "register 2 1 lazar root\n",
			Expected:    utils.MESSAGE.WrapSuccess("User registered in job #1 for Event #2 Baleinev 2023.\n"),
		},
		{
			Description: "Send register command with bad ids and receive error message",
			Input:       "register bad id lazar root\n",
			Expected:    utils.MESSAGE.Error.MustBeInteger,
		},
		{
			Description: "Send register command for inexistant event and receive error message",
			Input:       "register 1000 1 valentin root\n",
			Expected:    utils.MESSAGE.Error.EventNotFound,
		},
		{
			Description: "Send register command for inexistant job for the given event and receive error message",
			Input:       "register 2 1000 valentin root\n",
			Expected:    utils.MESSAGE.Error.JobNotFound,
		},
		{
			Description: "Send register command with bad credentials and receive receive error message",
			Input:       "register 2 1 jane rooot\n",
			Expected:    utils.MESSAGE.Error.AccessDenied,
		},
		{
			Description: "Send register command with user already in job and receive error message",
			Input:       "register 2 1 valentin root\n",
			Expected:    utils.MESSAGE.Error.AlreadyRegistered,
		},
		{
			Description: "Send register command for job already full and receive error message",
			Input:       "register 2 2 claude root\n",
			Expected:    utils.MESSAGE.Error.JobFull,
		},
		{
			Description: "Send register command as creator of event and receive error message",
			Input:       "register 2 1 john root\n",
			Expected:    utils.MESSAGE.Error.CreatorRegister,
		},
		{
			Description: "Send register command on closed event and receive error message",
			Input:       "register 1 1 lazar root\n",
			Expected:    utils.MESSAGE.Error.EventClosed,
		},
		{
			Description: "Send invalid register command and receive error message",
			Input:       "registeer bad input\n",
			Expected:    utils.MESSAGE.Error.InvalidCommand,
		},
	}
	testClient.Run(tests, t)
}

func Test_Commands_Concurrently(t *testing.T) {

	var message = utils.RED + "Closed" + utils.RESET + "\t#1 " + utils.BOLD + utils.CYAN + "Montreux Jazz 2022" + utils.RESET + " / Creator: claude\n\n" +
		utils.GREEN + "Open" + utils.RESET + "\t#2 " + utils.BOLD + utils.CYAN + "Baleinev 2023" + utils.RESET + " / Creator: john\n\n" +
		utils.GREEN + "Open" + utils.RESET + "\t#3 " + utils.BOLD + utils.CYAN + "Balélec 2023" + utils.RESET + " / Creator: jane\n"

	tests := []TestInput{
		{
			Description: "Send help command and receive help message",
			Input:       "help\n",
			Expected:    utils.MESSAGE.Help,
		},
		{
			Description: "Send show command and receive message",
			Input:       "show\n",
			Expected:    utils.MESSAGE.WrapEvent(message),
		},
	}

	runConcurrent(2, tests, t)
}
