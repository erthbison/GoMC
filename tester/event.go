package tester

type Event struct {
	Type    string
	Payload Message
}

func EventsEquals(a, b Event) bool {
	if a.Type != b.Type {
		return false
	}
	return a.Payload.Equals(b.Payload)
}
