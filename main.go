package main

import (
	"log"

	"DayOneCLI/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
