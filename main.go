package main

import "github.com/redhat-consulting-services/kustomize-validator/commands"

func main() {
	err := commands.RootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
