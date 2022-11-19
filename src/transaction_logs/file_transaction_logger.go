package transaction_logs

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
)

// FileTransactionLogger is a struct that satisfies the TransactionLogger
// interface, and writes logs to a file, with each log seperated by newlines
type FileTransactionLogger struct {
	events       chan<- Event // Write-only channel for sending events
	errors       <-chan error // Read-only channel for receiving errors
	lastSequence uint64       // The last recorded event sequence number
	file         *os.File     // The location of the transaction log
	wg           *sync.WaitGroup
}

// NewFileTransactionLogger is a constuctor for the FileTransactionLogger type,
// it takes a filename specifying where the log file is located, and
// it returns a TransactionLogger interface or any errors if they occur
func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	// Open the transaction log file for reading and writing.
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	return &FileTransactionLogger{file: file, wg: &sync.WaitGroup{}}, nil
}

// Run method initializes channels, listens for channel inputs and logs events accordingly
func (l *FileTransactionLogger) Run() {
	// Initialize the events channel
	events := make(chan Event, 16)
	l.events = events
	// Initialize the errors channel
	errs := make(chan error, 1)
	l.errors = errs
	// Run a goroutine to constantly handle new events coming over channels
	go func() {
		for e := range events {
			l.lastSequence++
			// Create the string to be written
			eventString := fmt.Sprintf("%d\t%d\t%s\t%s\n",
				l.lastSequence, e.EventType, e.Key, e.Value)
			// Write the string to the file
			_, err := l.file.Write([]byte(eventString))
			if err != nil {
				errs <- fmt.Errorf("cannot write to log file: %w", err)
			}
			l.wg.Done()
		}
	}()
}

// ReadEvents method parses an existing log file and creates an event for each line
// and broadcasts it over a read-only channel
func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	// Initialize scanner and output channels
	scanner := bufio.NewScanner(l.file) // Scanner for the logger to read the log file
	outEvent := make(chan Event)        // Unbuffered event channel to stream concurrent events
	outError := make(chan error, 1)     // Buffered error channel to stream concurrent errors
	restoredLines := 0                  // Used to check whether we restored from a txLog
	// Create concurrent process to parse the TxLog and refill the data
	go func() {
		var e Event // Event object to store data parsed from log
		// Close the channels when the goroutine ends
		defer close(outEvent)
		defer close(outError)
		// Go through each line in the txLog
		for scanner.Scan() {
			// Mark that we have restored data from the file
			restoredLines++
			// Parse through the line and create an event object
			line := scanner.Text()
			fmt.Sscanf(line, "%d\t%d\t%s\t%s", &e.Sequence, &e.EventType, &e.Key, &e.Value)
			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}
			outEvent <- e

		}
		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
		if restoredLines > 1 {
			log.Print("Restored vile store from transaction log")
		} else {
			log.Print("No transaction log found, creating new vile store")
		}
	}()

	return outEvent, outError
}

// WritePut method logs PUT events for the provided key:value pair as a line in a log file
func (l *FileTransactionLogger) WritePut(key, value string) {
	l.wg.Add(1)
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

// WriteDelete method logs DELETE events for the provided key as a line in a log file
func (l *FileTransactionLogger) WriteDelete(key string) {
	l.wg.Add(1)
	l.events <- Event{EventType: EventDelete, Key: key}
}

// Wait method waits for current threads to compelte
func (l *FileTransactionLogger) Wait() {
	l.wg.Wait()
}

// Close method gracefully shuts channels
func (l *FileTransactionLogger) Close() error {
	l.wg.Wait()
	if l.events != nil {
		close(l.events)
	}
	return l.file.Close()
}

// Err method returns any errors that have been read from the logger's error channel
func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}

// LastSequence method gets the last sequence of the txLog
func (l *FileTransactionLogger) LastSequence() uint64 {
	return l.lastSequence
}
