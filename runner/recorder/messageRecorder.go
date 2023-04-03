package recorder

type MessageRecorder interface {
	Subscribe() <-chan Message
}
