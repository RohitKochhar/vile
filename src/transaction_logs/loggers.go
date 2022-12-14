package transaction_logs

import (
	"fmt"
	"rohitsingh/vile/core"
)

// transaction_logs needs core access to write to store when loading
// information from the transaction logs

// TransactionLogger interface defines the required
// methods for a struct needed to serve as a transaction logger
type TransactionLogger interface {
	WriteDelete(key string)                   // WriteDelete logs DELETE events for a provided key
	WritePut(key, value string)               // WritePut logs PUT events for a provided key:value pair
	Err() <-chan error                        // Err returns any errors that have been read from the logger's error channel
	ReadEvents() (<-chan Event, <-chan error) // ReadEvents parses the logfile and creates an event for each line
	Run()                                     // Run starts the logger, accepts new events put over channels and writes them to the log
	Wait()                                    // Wait waits for any concurrent threads to complete before unblocking
	Close() error                             // Close gracefully closes the TransactionLogger
	LastSequence() uint64                     // Returns the last sequence in a file txlog
}

// initializeTransactionLog creates a TransactionLogger object, watches for events and logs
// them accordingly
// If filepath is an empty string, a postgres db is created instead
func InitializeTransactionLog(filepath string) (TransactionLogger, error) {
	var transact TransactionLogger
	var err error
	if filepath == "" {
		transact, err = NewPostgresTransactionLogger(PostgresDBConfig{
			host:     "host.docker.internal",
			dbName:   "vile",
			user:     "test",
			password: "password",
		})
	} else {
		transact, err = NewFileTransactionLogger(filepath)
	}
	if err != nil {
		return nil, fmt.Errorf("unexpected error while creating event logger: %w", err)
	}
	events, errors := transact.ReadEvents()
	e, ok := Event{}, true
	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.EventType {
			case EventDelete:
				err = core.Delete(e.Key)
			case EventPut:
				err = core.Put(e.Key, e.Value)
			}
		}
	}
	transact.Run()
	return transact, err
}
