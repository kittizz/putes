package cmd

import (
	"log"
)

// Execute executes the commands.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
