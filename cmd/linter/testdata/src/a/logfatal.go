package a

import "log"

func testLogFatal() {
	log.Fatal("test") // want "call to log.Fatal outside main.main is forbidden"
}

func testLogFatalf() {
	log.Fatalf("test %s", "x") // want "call to log.Fatalf outside main.main is forbidden"
}
