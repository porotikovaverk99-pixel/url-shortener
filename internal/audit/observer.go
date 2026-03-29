package audit

type Observer interface {
	Send(event Event) error
}
