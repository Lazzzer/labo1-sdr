package client

import (
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/Lazzzer/labo1-sdr/server"
	"github.com/Lazzzer/labo1-sdr/utils"
)

type TestInput struct {
	Description string
	Input       []byte
	Expected    []byte
}

type TestClient struct {
	Config utils.Config
}

func (tc *TestClient) Run(tests []TestInput, t *testing.T) {
	conn, err := net.Dial("tcp", tc.Config.Host+":"+strconv.Itoa(tc.Config.Port))

	if err != nil {
		t.Error("Could not connect to server", err)
	} else {
		fmt.Println("Welcome! Please enter a command.\nType 'help' for a list of commands.")
	}
	defer conn.Close()

	for _, test := range tests {
		fmt.Println(test.Description)
		conn.Write(test.Input)
		buf := make([]byte, 1024)
		conn.Read(buf)
		if string(buf) != string(test.Expected) {
			t.Errorf("Expected %s, got %s", string(test.Expected), string(buf))
		}
	}
}

func TestClient_Can_Connect(t *testing.T) {

	server := server.Server{Config: utils.Config{Port: 8081, Debug: false}}
	go server.Run()

	testClient := TestClient{Config: utils.Config{Host: "localhost", Port: 8081}}

	tests := []TestInput{
		{
			Description: "Client can connect and get welcome message",
			Input:       []byte(""),
			Expected:    []byte("Welcome! Please enter a command.\nType 'help' for a list of commands."),
		},
	}

	testClient.Run(tests, t)
}
