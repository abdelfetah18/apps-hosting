package main

import (
	"go_registry/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
