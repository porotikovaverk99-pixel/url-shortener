package main

import (
	"log"
	"os"
)

func main() {
	// allowed in main.main
	log.Fatal("ok")
	os.Exit(0)
	panic("allowed in main")
}

func notMain() {
	log.Fatal("forbidden") // want "call to log.Fatal outside main.main is forbidden"
	os.Exit(1)             // want "call to os.Exit outside main.main is forbidden"
	panic("forbidden")     // want "use of panic is forbidden"
}
