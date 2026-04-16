package a

import (
	"log"
	"os"
	"runtime"
)

func main() {
}

func someFunc() {
	panic("forbidden")     // want "use of panic is forbidden"
	log.Fatal("forbidden") // want "call to log.Fatal outside main.main is forbidden"
	os.Exit(1)             // want "call to os.Exit outside main.main is forbidden"
	runtime.Goexit()       // want "call to runtime.Goexit outside main.main is forbidden"
}
