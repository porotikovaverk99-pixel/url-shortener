package a

func testPanic() {
	panic("test") // want "use of panic is forbidden"
}
