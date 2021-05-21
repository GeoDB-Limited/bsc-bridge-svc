package data

type Status string

const (
	StatusSent     Status = "sent"
	StatusNotSent  Status = "not_sent"
	StatusFailed   Status = "failed"
	StatusRefunded Status = "refunded"
)
