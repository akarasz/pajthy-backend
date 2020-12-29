package event

type Role int

const (
	Voter Role = iota
	Controller
)

type Type string

const (
	Enabled            = "enabled"
	Disabled           = "disabled"
	Reset              = "reset"
	ParticipantsChange = "participants-change"
	Vote               = "vote"
	Done               = "done"
)

type Payload struct {
	Kind Type
	Data interface{}
}

type Event interface {
	Emit(sessionID string, r Role, t Type, body interface{})
	Subscribe(sessionID string, r Role, ws interface{}) (chan *Payload, error)
	Unsubscribe(sessionID string, ws interface{}) error
}
