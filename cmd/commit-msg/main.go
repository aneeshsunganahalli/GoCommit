package main

import (
	"log"

	cmd "github.com/dfanso/commit-msg/cmd/cli"
	"github.com/dfanso/commit-msg/cmd/cli/store"
)

// main is the entry point of the commit message generator
func main() {

	//Initializes the OS credential manager
	KeyRing, err := store.KeyringInit()
	if err != nil {
		log.Fatalf("Failed to initilize Keyring store: %v", err)
	}
	cmd.StoreInit(KeyRing) //Passes StoreMethods instance to root
	cmd.Execute()
}
