package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/urfave/cli/v2"
)

const DefaultPrompt = "> "

type Config struct {
	DataDir string // Data directory to store the console history at
	DocRoot string // Filesystem path from where to load JavaScript files from
	//Client   *rpc.Client  // RPC client to execute Ethereum requests through
	Prompt   string        // Input prompt prefix string (defaults to DefaultPrompt)
	Prompter *bufio.Reader // Input prompter to allow interactive user feedback (defaults to TerminalPrompter)
	Printer  io.Writer     // Output writer to serialize any display strings to (defaults to os.Stdout)
	Preload  []string      // Absolute paths to JavaScript files to preload
}

type Console struct {
	prompter *bufio.Reader // Input prompter to allow interactive user feedback (defaults to TerminalPrompter)
	printer  io.Writer     // Output writer to serialize any display strings to
	config   Config
}

func New(config Config) (*Console, error) {
	// Handle unset config values gracefully
	config.Prompter = bufio.NewReader(os.Stdin)
	config.Prompt = DefaultPrompt
	config.Printer = colorable.NewColorableStdout()

	// Initialize the console and return
	console := &Console{
		prompter: config.Prompter,
		printer:  config.Printer,
	}

	return console, nil
}

// Welcome show summary of current Database and some metadata about the
// console's available modules.
func (c *Console) Welcome() {
	message := "Welcome to the GoDBLedger Reporter console!\n\n"
	message += "Database File: " + c.config.DataDir + "\n"

	fmt.Fprintln(c.printer, message)
}

// Evaluate executes code and pretty prints the result to the specified output
// stream.
func (c *Console) Evaluate(statement string) error {
	switch strings.TrimSpace(statement) {
	case "quit":
		os.Exit(1)
	case "tb":
		fmt.Println("Today is 5th. Clean your house.")
	case "gl":
		fmt.Println("Today is 5th. Clean your house.")
	default:
		fmt.Println("No information available for that day.")
	}

	return nil
}

// PromptInput displays the given prompt to the user and requests some textual
// data to be entered, returning the input of the user.
func (c *Console) PromptInput(prompt string) (string, error) {
	fmt.Print(prompt)
	//prompt = ""
	defer fmt.Println()

	return c.prompter.ReadString('\n')
}

func reporterConsole(c *cli.Context) error {
	console, _ := New(Config{})
	console.Welcome()

	something, _ := console.PromptInput("~:")
	console.Evaluate(something)

	return nil
}
