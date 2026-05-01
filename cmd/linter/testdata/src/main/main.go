package main

import (
	"log"
	"os"
	"runtime"
)

func init() {
	panic("allowed in init")
}

func main() {
	panic("allowed in main")

	log.Fatal("allowed in main")

	os.Exit(0)

	runtime.Goexit()

	func() {
		panic("allowed in main nested")
	}()
}

func notMain() {
	panic("forbidden outside main") // want "use of panic is forbidden"

	log.Fatal("forbidden outside main") // want "call to log.Fatal outside main.main is forbidden"

	os.Exit(1) // want "call to os.Exit outside main.main is forbidden"

	runtime.Goexit() // want "call to runtime.Goexit outside main.main is forbidden"
}

func someFunc() {
	defer func() {
		panic("forbidden in defer outside main") // want "use of panic is forbidden"
	}()

	go func() {
		panic("forbidden in goroutine outside main") // want "use of panic is forbidden"
	}()
}
