package utils

var MESSAGE = struct {
	Success success
	Error   error
	Help    string
}{
	Success: success{
		Created: "Job created successfully.\n",
	},
	Error: error{
		InvalidCommand: "Error: Invalid command. Type 'help' for a list of commands.\n",
	},
	Help: help,
}

type success = struct {
	Created string
}

type error = struct {
	InvalidCommand string
}

var help = YELLOW +
	"\n===================== 💡 HELP 💡 =============================\n\n" + RESET +
	"ℹ️ Arguments with brackets [] are optional.\n\n" +
	"ℹ️ Commands with \"🔒\" need credentials (arguments in double brackets [[]]) to be used.\n" +
	"If you are using the client, you will have a prompt for them.\n" +
	"Otherwise, you have to put your credentials directly in the command.\n\n" +
	YELLOW + "Commands list:" + RESET + "\n\n" +
	"# Display help and list all commands\n" +
	GREEN + "help" + RESET + "\n\n" +
	"# 🔒 Create an event with a list of jobs and its number of volunteers needed\n" +
	GREEN + "create" + RESET + " <eventName> <jobName1> <nbVolunteer1> [<jobName2> <nbVolunteer2>...] [[<username> <password>]]\n\n" +
	"# 🔒 Close an event\n" +
	GREEN + "close" + RESET + " <idEvent> [[<username> <password>]]\n\n" +
	"# 🔒 Register as a volunteer to a job\n" +
	GREEN + "register" + RESET + " <idEvent> <idJob> [[<username> <password>]]\n\n" +
	"# Show all events. If the id is specified, show the event with all its jobs instead\n" +
	GREEN + "show" + RESET + " [<idEvent>]\n\n" +
	"# Show the distribution of volunteers from each job of an event\n" +
	GREEN + "jobs" + RESET + " <idEvent>\n\n" +
	"# Quit the program\n" +
	GREEN + "quit" + RESET + "\n\n" +
	YELLOW + "=============================================================" + RESET + "\n\n"

// Success messages

// Error messages
