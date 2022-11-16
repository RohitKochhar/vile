package transaction_logs

import (
	"bufio"
	"fmt"
	"os"
	// transaction_logs needs core access to write to store when loading
	// information from the transaction logs
)

// TransactionLogger interface defines the required
// methods for a struct needed to serve as a transaction logger
type TransactionLogger interface {
	WriteDelete(key string)                   // WriteDelete logs DELETE events for a provided key
	WritePut(key, value string)               // WritePut logs PUT events for a provided key:value pair
	Err() <-chan error                        // Err returns any errors that have been read from the logger's error channel
	ReadEvents() (<-chan Event, <-chan error) // ReadEvents parses the logfile and creates an event for each line
	Run()                                     // Run starts the logger, accepts new events put over channels and writes them to the log
}

// FileTransactionLogger is a struct that satisfies the TransactionLogger
// interface, and writes logs to a file, with each log seperated by newlines
type FileTransactionLogger struct {
	events       chan<- Event // Write-only channel for sending events
	errors       <-chan error // Read-only channel for receiving errors
	lastSequence uint64       // The last recorded event sequence number
	file         *os.File     // The location of the transaction log
}

// NewFileTransactionLogger is a constuctor for the FileTransactionLogger type,
// it takes a filename specifying where the log file is located, and
// it returns a TransactionLogger object or any errors if they occur
func NewFileTransactionLogger(filename string) (*FileTransactionLogger, error) {
	var err error
	var l FileTransactionLogger

	// Open the transaction log file for reading and writing.
	l.file, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}
	l.events = make(chan Event, 16)
	l.errors = make(chan error, 1)

	return &l, nil
}

// Run method initializes channels, listens for channel inputs and logs events accordingly
func (l *FileTransactionLogger) Run() {
	// Initialize the events channel
	events := make(chan Event, 16)
	l.events = events

	// Initialize the errors channel
	errs := make(chan error, 1)
	l.errors = errs
	// Run a goroutine to constantly handling new events coming over channels
	go func() {
		for e := range events {
			l.lastSequence++
			// Create the string to be written
			eventString := fmt.Sprintf("%d\t%d\t%s\t%s\n",
				l.lastSequence, e.EventType, e.Key, e.Value)
			// Write the string to the file
			_, err := l.file.Write([]byte(eventString))

			// _, err := fmt.Fprintf(
			// 	l.file,
			// 	"%d\t%d\t%s\t%s\n",
			// 	l.lastSequence, e.EventType, e.Key, e.Value)

			if err != nil {
				errs <- fmt.Errorf("cannot write to log file: %w", err)
			}
		}
	}()
}

// ReadEvents method parses an existing log file and creates an event for each line
// and broadcasts it over a read-only channel
func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file) // Scanner for the logger to read the log file
	outEvent := make(chan Event)        // Unbuffered event channel to stream concurrent events
	outError := make(chan error, 1)     // Buffered error channel to stream concurrent errors
	go func() {
		var e Event // Event object to store data parsed from log
		// Close the channels when the goroutine ends
		defer close(outEvent)
		defer close(outError)
		for scanner.Scan() {
			line := scanner.Text()
			if _, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s",
				&e.Sequence, &e.EventType, &e.Key, &e.Value); err != nil {
				outError <- fmt.Errorf("unexpected error while parsing input: %w", err)
				return
			}
			// Check that the sequence numbers are in increasing order as we would expect
			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}
			// Update the last event of the logger and transmit the event over the channel
			l.lastSequence = e.Sequence
			outEvent <- e
		}
		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()
	return outEvent, outError
}

// WritePut method logs PUT events for the provided key:value pair as a line in a log file
func (l *FileTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

// WriteDelete method logs DELETE events for the provided key as a line in a log file
func (l *FileTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

// Err method returns any errors that have been read from the logger's error channel
func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}
