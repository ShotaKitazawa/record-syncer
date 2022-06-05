package models

type WatchResponse struct {
	Type  EventType
	Key   []byte
	Value []byte
}

type EventType int32

const (
	PUT    EventType = 0
	DELETE EventType = 1
)
