package main

import (
	"fmt"

	"github.com/urfave/cli"
)

const DefaultPrompt = "> "

type Console struct {
	printer io.Writer // Output writer to serialize any display strings to
}

func New(config Config) (*Console, error) {
	// Handle unset config values gracefully
	config.Prompter = Stdin
	config.Printer = colorable.NewColorableStdout()

	// Initialize the console and return
	console := &Console{
		prompter: config.Prompter,
		printer:  config.Printer,
	}
}

// Welcome show summary of current Database and some metadata about the
// console's available modules.
func (c *Console) Welcome() {
	message := "Welcome to the GoDBLedger Reporter console!\n\n"
	message += "Database File: " + defaultDBName + "\n"

	fmt.Fprintln(c.printer, message)
}

// PromptInput displays the given prompt to the user and requests some textual
// data to be entered, returning the input of the user.
func (c *Console) PromptInput(prompt string) (string, error) {
	if p.supported {
		p.rawMode.ApplyMode()
		defer p.normalMode.ApplyMode()
	} else {
		// liner tries to be smart about printing the prompt
		// and doesn't print anything if input is redirected.
		// Un-smart it by printing the prompt always.
		fmt.Print(prompt)
		prompt = ""
		defer fmt.Println()
	}
	return p.State.Prompt(prompt)
}

func reporterConsole(c *cli.Context) error {
	fmt.Printf("Hello %q", c.Args().Get(0))
	return nil
}
