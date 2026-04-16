package init

func init() {
	panic("allowed in init")
}

func main() {
}

func someFunc() {
	panic("forbidden") // want "use of panic is forbidden"
}
