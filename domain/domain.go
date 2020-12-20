package domain

type Vote struct {
	Participant string
	Choice      string
}

type Session struct {
	Choices      []string
	Participants []string
	Votes        []Vote
	Open         bool
}
