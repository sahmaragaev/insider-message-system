package messagestatus

type MessageStatus string

const (
	Pending MessageStatus = "pending"
	Sent    MessageStatus = "sent"
	Failed  MessageStatus = "failed"
)

func (m MessageStatus) String() string {
	return string(m)
}

func (m MessageStatus) IsValid() bool {
	switch m {
	case Pending, Sent, Failed:
		return true
	default:
		return false
	}
}

func FromString(status string) MessageStatus {
	ms := MessageStatus(status)
	if ms.IsValid() {
		return ms
	}
	return Pending
}
