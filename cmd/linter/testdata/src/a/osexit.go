package a

import "os"

func testOsExit() {
	os.Exit(1) // want "call to os.Exit outside main.main is forbidden"
}
