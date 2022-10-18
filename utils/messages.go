package utils

type Message struct {
	Error      errorMessage
	Title      string
	Goodbye    string
	Help       string
	LoginStart string
	LoginEnd   string
}

type errorMessage = struct {
	InvalidCommand      string
	InvalidNbArgs       string
	AccessDenied        string
	MustBeInteger       string
	EventNotFound       string
	EventClosed         string
	JobNotFound         string
	NotCreator          string
	AlreadyClosed       string
	IdEventNotMatchJob  string
	CreatorRegister     string
	JobFull             string
	AlreadyRegistered   string
	NbVolunteersInteger string
}

var MESSAGE = Message{
	Error: errorMessage{
		InvalidCommand:      wrapError("Invalid command. Type 'help' for a list of commands.\n"),
		InvalidNbArgs:       wrapError("Invalid number of arguments. Type 'help' for more information.\n"),
		AccessDenied:        wrapError("Access denied.\n"),
		MustBeInteger:       wrapError("Id must be an integer.\n"),
		EventNotFound:       wrapError("Event not found with given id.\n"),
		EventClosed:         wrapError("Event is closed.\n"),
		JobNotFound:         wrapError("Job not found with given id.\n"),
		NotCreator:          wrapError("Only the creator of the event can close it.\n"),
		AlreadyClosed:       wrapError("Event is already closed.\n"),
		IdEventNotMatchJob:  wrapError("Given event id does not match id in job.\n"),
		CreatorRegister:     wrapError("Creator of the event cannot register for a job.\n"),
		JobFull:             wrapError("Job is already full.\n"),
		AlreadyRegistered:   wrapError("User is already registered in this job.\n"),
		NbVolunteersInteger: wrapError("Number of volunteers must be a positive integer.\n"),
	},
	Title:      title,
	Goodbye:    goodbye,
	Help:       help,
	LoginStart: loginStart,
	LoginEnd:   loginEnd,
}

func (m *Message) WrapSuccess(message string) string {
	success := GREEN + "\n===================== ✅ SUCCESS ✅ ==========================\n\n" + RESET
	success += message + "\n"
	success += GREEN + "==============================================================" + RESET + "\n\n"
	return success
}

func (m *Message) WrapEvent(message string) string {
	event := CYAN + "\n====================== 📅 EVENT 📅 ===========================\n\n" + RESET
	event += message + "\n"
	event += CYAN + "==============================================================" + RESET + "\n\n"
	return event
}

func wrapError(message string) string {
	err := RED + "\n===================== ❌ ERROR ❌ ============================\n\n" + RESET
	err += message + "\n"
	err += RED + "==============================================================" + RESET + "\n\n"

	return err
}

var title = BOLD + YELLOW +
	"  _____                 _     __  __                                   \n" +
	" | ____|_   _____ _ __ | |_  |  \\/  | __ _ _ __   __ _  __ _  ___ _ __ \n" +
	" |  _| \\ \\ / / _ \\ '_ \\| __| | |\\/| |/ _` | '_ \\ / _` |/ _` |/ _ \\ '__|\n" +
	" | |___ \\ V /  __/ | | | |_  | |  | | (_| | | | | (_| | (_| |  __/ |   \n" +
	" |_____| \\_/ \\___|_| |_|\\__| |_|  |_|\\__,_|_| |_|\\__,_|\\__, |\\___|_|   \n" +
	"                                                       |___/           " + RESET + "\n" +
	"Labo 1 SDR - Jonathan Friedli & Lazar Pavicevic\n\n" +
	"Welcome! Please enter a command.\n" +
	"💡" + YELLOW + "Type 'help' for a list of available commands." + RESET + "\n"

var goodbye = "\nThank you for using Event Manager!\n" + BOLD + YELLOW +
	"   ______                __   __               __\n" +
	"  / ____/___  ____  ____/ /  / /_  __  _____  / /\n" +
	" / / __/ __ \\/ __ \\/ __  /  / __ \\/ / / / _ \\/ / \n" +
	"/ /_/ / /_/ / /_/ / /_/ /  / /_/ / /_/ /  __/_/  \n" +
	"\\____/\\____/\\____/\\__,_/  /_.___/\\__, /\\___(_)   \n" +
	"                                /____/           " + RESET

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
	YELLOW + "==============================================================" + RESET + "\n\n"

var loginStart = ORANGE +
	"\n====================== 🔑 LOGIN 🔑 ===========================" + RESET + "\n"

var loginEnd = ORANGE +
	"\n\n==============================================================" + RESET
