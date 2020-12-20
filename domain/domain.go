package domain

type Vote struct {
	Participant string
	Choice      string
}

type Session struct {
	Choices      map[string]interface{}
	Participants map[string]interface{}
	Votes        []*Vote
	Open         bool
}

func NewSession() *Session {
	return &Session{
		Choices:      map[string]interface{}{},
		Participants: map[string]interface{}{},
		Votes:        []*Vote{},
		Open:         false,
	}
}
