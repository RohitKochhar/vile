package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	// server needs transaction_logs access to record
	// HTTP request history in the transaction log
	"rohitsingh/vile/transaction_logs"
)

// tester is a generic test interface that can be used
// to allow functions to accept either testing.T or
// testing.B objects
type tester interface {
	// Helper marks the calling function as a test helper function.
	// When printing file and line information, that function will be skipped.
	// Helper may be called simultaneously from multiple goroutines.
	Helper()
	// Fatal is equivalent to Log followed by FailNow.
	Fatal(args ...any)
	// Fatalf is equivalent to Logf followed by FailNow.
	Fatalf(format string, args ...any)
}

// setupAPI is a helper function that sets up
// the API for the tests, providing a cleanup function too
func setupAPI(t tester) (string, func()) {
	t.Helper() // Mark the function as test helper
	ts := httptest.NewServer(NewMux())
	// Create a temp transaction file
	file, err := os.CreateTemp("", "transaction.log")
	if err != nil {
		t.Fatal(err)
	}
	transact, err = transaction_logs.InitializeTransactionLog(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return ts.URL, func() {
		ts.Close()
		os.Remove(file.Name())
	}
}

// getHelper wraps the Get function in additional logic to
// assist with testing ease and clarity
func getHelper(t tester, getUrl string, expBody string, expCode int) (r *http.Response) {
	r, err := http.Get(getUrl)
	if err != nil {
		t.Fatalf("error while sending GET request: %q", err)
	}
	// Check if the return code is what we expected
	if r.StatusCode != expCode {
		t.Fatalf("Expected %q, got %q.", http.StatusText(expCode),
			http.StatusText(r.StatusCode))
	}
	defer r.Body.Close()
	// We might not be expecting content
	if expBody != "" || expCode == http.StatusNotFound {
		// The result of GET from the server should always be in
		// plain text
		if !strings.Contains(r.Header.Get("content-Type"), "text/plain") {
			t.Fatalf("unsupported Content-Type: %q", r.Header.Get("Content-Type"))
		}
		// Check that we have the content that we expected
		var body []byte
		if body, err = io.ReadAll(r.Body); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), expBody) {
			t.Fatalf("Expected %q, got %q.", expBody, string(body))
		}
	}

	return r
}

// putHelper wraps the Put function in additional logic to
// assist with testing ease and clarity
func putHelper(t tester, putUrl string, val string, expCode int) *http.Response {
	req, err := http.NewRequest(
		http.MethodPut,
		putUrl,
		bytes.NewBuffer([]byte(val)),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != expCode {
		t.Fatalf("Expected %q, got %q.", http.StatusText(expCode),
			http.StatusText(resp.StatusCode))
	}
	return resp
}

// delHelper wraps the Delete function in additional logic to
// assist with testing ease and clarity
func delHelper(t tester, delUrl string, expCode int) *http.Response {
	req, err := http.NewRequest(
		http.MethodDelete,
		delUrl,
		bytes.NewBuffer([]byte("")),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %q, got %q.", http.StatusText(http.StatusOK),
			http.StatusText(resp.StatusCode))
	}
	return resp
}

// TestGet tests HTTP get method on the server's root
func TestGet(t *testing.T) {
	testCases := []struct {
		name       string // Name of test
		path       string // URL path to GET
		expCode    int    // Expected HTTP return code
		expContent string // Expected result
	}{
		{
			"GetRoot", "/", http.StatusOK, "Check now hey! This is vile, man!",
		},
		{
			"NotFound", "/badpth", http.StatusNotFound, "",
		},
	}
	// Create a test server hosting the API
	url, cleanup := setupAPI(t)
	defer cleanup()
	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use the get helper to check get success
			_ = getHelper(t, url+tc.path, tc.expContent, tc.expCode)
		})
	}
}

func TestPut(t *testing.T) {
	testCases := []struct {
		name    string // Name of test
		path    string // Path to use
		key     string // key to PUT
		value   string // Value to PUT
		expCode int    // Expected result
	}{
		{
			"TestShortPut", "/", "testKey", "val", http.StatusCreated,
		},
		{
			"TestLongPut", "/v1/key/", "testKey", "val", http.StatusCreated,
		},
	}
	// Create a test server hosting the API
	url, cleanup := setupAPI(t)
	defer cleanup()
	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use the put helper to check PUT success
			_ = putHelper(t, url+tc.path+tc.key, tc.value, tc.expCode)
		})
	}
}

func TestIntegration(t *testing.T) {
	// Set up the test server
	url, cleanup := setupAPI(t)
	defer cleanup()
	// Put a value in the store
	key := "testKey"
	val := "testVal"
	paths := []string{"/", "/v1/key/"}
	for _, p := range paths {
		path := url + p + key
		// PUT the value in the store
		_ = putHelper(t, path, val, http.StatusCreated)
		// GET the stored value
		_ = getHelper(t, path, val, http.StatusOK)
		// DELETE the value
		_ = delHelper(t, path, http.StatusOK)
		// GET the value, but expect to fail since it was deleted.
		_ = getHelper(t, path, "", http.StatusNotFound)
	}
}
