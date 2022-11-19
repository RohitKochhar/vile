package transaction_logs

// Event type contains the information to be logged by
// a TransactionLogger interface
type Event struct {
	Sequence  uint64    // Unique record ID
	EventType EventType // Action taken in event
	Key       string    // Key affected by this event
	Value     string    // Value PUT by this event (only for PUTs)
}

// EventType type assigns a byte-value to each possible event
// for consistency across functions
type EventType byte

const (
	_                     = iota
	EventDelete EventType = iota // EventType corresponding to a DELETE action
	EventPut                     // Eventype corresponding to a PUT action
)
