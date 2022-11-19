package transaction_logs

import (
	"fmt"
	"os"
	"testing"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func TestCreateFileLogger(t *testing.T) {
	// Create a temp file to store log
	const filename = "/tmp/file-logger-test.log"
	defer os.Remove(filename)
	// Create the file logger
	logger, err := NewFileTransactionLogger(filename)
	if err != nil {
		t.Fatalf("unexpected error while creating file transaction logger: %q", err)
	}
	if logger == nil {
		t.Fatal("unexpected nil value returned for file transaction logger")
	}
	if !fileExists(filename) {
		t.Fatalf("transaction log file (%s) does not exist", filename)
	}
}

func evaluateLastSequence(t *testing.T, el TransactionLogger, expected uint64) {
	if ls := el.LastSequence(); ls != expected {
		t.Errorf("Last sequence mismatch (expected %d; got %d)", expected, ls)
	} else {
		t.Logf("Last sequence agrees with expectations (expected %d; got %d)", expected, ls)
	}
}

func TestWritePut(t *testing.T) {
	// Create a transaction logger
	// Create a temp file to store log
	const filename = "/tmp/file-logger-test.log"
	defer os.Remove(filename)
	// Create the file logger
	logger, err := NewFileTransactionLogger(filename)
	if err != nil {
		t.Fatalf("unexpected error while creating file transaction logger: %q", err)
	}
	logger.Run()
	defer logger.Close()
	// Add some values to the log
	evaluateLastSequence(t, logger, 0)
	logger.WritePut("key", "val")
	logger.Wait()
	evaluateLastSequence(t, logger, 1)
	for i := 0; i < 10; i++ {
		logger.WritePut(fmt.Sprintf("key%d", i), "someval")
	}
	logger.Wait()
	evaluateLastSequence(t, logger, 11)
}

func TestWriteAppend(t *testing.T) {
	const filename = "/tmp/write-append.txt"
	defer os.Remove(filename)

	tl, err := NewFileTransactionLogger(filename)
	if err != nil {
		t.Error(err)
	}
	tl.Run()
	defer tl.Close()

	chev, cherr := tl.ReadEvents()
	for e := range chev {
		t.Log(e)
	}
	err = <-cherr
	if err != nil {
		t.Error(err)
	}

	tl.WritePut("my-key", "my-value")
	tl.WritePut("my-key", "my-value2")
	tl.Wait()

	tl2, err := NewFileTransactionLogger(filename)
	if err != nil {
		t.Error(err)
	}
	tl2.Run()
	defer tl2.Close()

	chev, cherr = tl2.ReadEvents()
	for e := range chev {
		t.Log(e)
	}
	err = <-cherr
	if err != nil {
		t.Error(err)
	}

	tl2.WritePut("my-key", "my-value3")
	tl2.WritePut("my-key2", "my-value4")
	tl2.Wait()

	if tl2.LastSequence() != 4 {
		t.Errorf("Last sequence mismatch (expected 4; got %d)", tl2.LastSequence())
	}
}
