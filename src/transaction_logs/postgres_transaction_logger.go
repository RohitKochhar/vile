package transaction_logs

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

// PostgresTransactionLogger is a struct the satisfies the TransactionLogger
// interface, and writes logs to a Postgres database
type PostgresTransactionLogger struct {
	events chan<- Event   // Write-only channel for sending events
	errors <-chan error   // Read-only channel for receiving errors
	db     *sql.DB        // Database access interface
	wg     sync.WaitGroup // Wait-group for concurrency
}

// PostgresDBConfig is a type containing information to configure the postgres db
type PostgresDBConfig struct {
	dbName   string // Name of the database
	host     string // Hostname where the database is hosted
	user     string // DB user
	password string // DB password
}

// NewPostgresTransactionLogger is a constructor for the PostgresTransactionLogger type,
// It takes PostgresDBConfig object containing db configuration information
// It returns a TransactionLogger interface or any errors if they occur
func NewPostgresTransactionLogger(c PostgresDBConfig) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable",
		c.host, c.dbName, c.user, c.password,
	)
	// Open connection to database
	log.Println("Attempting to open connection to postgres database...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error while opening database: %q", err)
	}
	// Test database connection
	err = db.Ping()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error while testing database connection: %q", err)
	}
	log.Println("Successfully connected to postgres database")
	logger := &PostgresTransactionLogger{db: db}
	// Check that the table exists
	exists, err := logger.verifyTableExists()
	if err != nil {
		return nil, fmt.Errorf("error while verifying table exists: %q", err)
	}
	if !exists {
		if err = logger.createTable(); err != nil {
			return nil, fmt.Errorf("error while creating table: %q", err)
		}
	}
	return logger, nil
}

// Run method initializes channels, listens for inputs and logs events to db
func (l *PostgresTransactionLogger) Run() {
	// Initialize events channel
	events := make(chan Event, 16)
	l.events = events
	// Initialize error channel
	errs := make(chan error, 1)
	l.errors = errs
	// Run a goroutine to constantly handle new events coming over channels
	go func() {
		query := `INSERT INTO transactions
			(event_type, key, value)
			VALUES ($1, $2, $3)`
		for e := range events {
			_, err := l.db.Exec(
				query,
				e.EventType, e.Key, e.Value,
			)
			if err != nil {
				errs <- err
			}
		}
	}()
}

// ReadEvents method parses an existing postgres db and loads prior events into
// store
func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)    // Unbuffered event channel to stream concurrent events
	outError := make(chan error, 1) // Buffered error channel to stream concurrent errors
	go func() {
		// Close the channels when the goroutine ends
		defer close(outEvent)
		defer close(outError)
		query := `SELECT sequence, event_type, key, value FROM transactions
				ORDER BY sequence`
		rows, err := l.db.Query(query)
		if err != nil {
			outError <- fmt.Errorf("error while running sql query: %q", err)
			return
		}
		defer rows.Close()
		e := Event{}
		for rows.Next() {
			err = rows.Scan(
				&e.Sequence,
				&e.EventType,
				&e.Key,
				&e.Value,
			)
			if err != nil {
				outError <- fmt.Errorf("error while reading row: %q", err)
				return
			}
			outEvent <- e
		}
		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("error while reading transaction log: %q", err)
		}
	}()
	return outEvent, outError
}

// WritePut method logs PUT events for the provided key:value pair to the postgres db
func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

// WriteDelete method logs DELETE events for the provided key to the postgres db
func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

// Err method returns any errors that have been read from the logger's error channel
func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}

// Wait method blocks until all concurrent threads have completed
func (l *PostgresTransactionLogger) Wait() {
	l.wg.Wait()
}

// Close method gracefully closes transactionlogger
func (l *PostgresTransactionLogger) Close() error {
	l.wg.Wait()

	if l.events != nil {
		close(l.events)
	}

	return l.db.Close()
}

// verifyTableExists
func (l *PostgresTransactionLogger) verifyTableExists() (bool, error) {
	const table = "transactions"

	var result string

	rows, err := l.db.Query(fmt.Sprintf("SELECT to_regclass('public.%s');", table))

	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() && result != table {
		rows.Scan(&result)
	}

	return result == table, rows.Err()
}

// createTable method
func (l *PostgresTransactionLogger) createTable() error {
	var err error

	createQuery := `CREATE TABLE transactions (
		sequence      BIGSERIAL PRIMARY KEY,
		event_type    SMALLINT,
		key 		  TEXT,
		value         TEXT
	  );`

	_, err = l.db.Exec(createQuery)
	if err != nil {
		return err
	}

	return nil
}

func (l *PostgresTransactionLogger) LastSequence() uint64 {
	return 0
}
