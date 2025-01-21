package main

import "github.com/leonsteinhaeuser/kustomize-validator/commands"

func main() {
	err := commands.RootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
