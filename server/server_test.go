package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// setupAPI is a helper function that sets up
// the API for the tests, providing a cleanup function too
func setupAPI(t *testing.T) (string, func()) {
	t.Helper() // Mark the function as test helper
	ts := httptest.NewServer(NewMux())
	return ts.URL, func() {
		ts.Close()
	}
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
			var (
				body []byte
				err  error
			)
			r, err := http.Get(url + tc.path)
			if err != nil {
				t.Error(err)
			}
			defer r.Body.Close()
			if r.StatusCode != tc.expCode {
				t.Fatalf("Expected %q, got %q.", http.StatusText(tc.expCode),
					http.StatusText(r.StatusCode))
			}
			switch {
			case strings.Contains(r.Header.Get("Content-Type"), "text/plain"):
				if body, err = io.ReadAll(r.Body); err != nil {
					t.Error(err)
				}
				if !strings.Contains(string(body), tc.expContent) {
					t.Errorf("Expected %q, got %q.", tc.expContent,
						string(body))
				}
			default:
				t.Fatalf("Unsupported Content-Type: %q", r.Header.Get("Content-Type"))
			}

		})
	}
}

func TestPut(t *testing.T) {
	testCases := []struct {
		name    string // Name of test
		key     string // URL path to PUT
		value   string // Value to PUT
		expCode int    // Expected result
	}{
		{
			"TestSimplePut", "testKey", "val", http.StatusCreated,
		},
	}
	// Create a test server hosting the API
	url, cleanup := setupAPI(t)
	defer cleanup()
	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				err error
			)
			path := url + "/v1/key/" + tc.key
			req, err := http.NewRequest(
				http.MethodPut,
				path,
				bytes.NewBuffer([]byte(tc.value)),
			)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "text/plain")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != tc.expCode {
				t.Fatalf("Expected %q, got %q.", http.StatusText(tc.expCode),
					http.StatusText(resp.StatusCode))
			}
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
	path := url + "/v1/key/" + key
	req, err := http.NewRequest(
		http.MethodPut,
		path,
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
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected %q, got %q.", http.StatusText(http.StatusCreated),
			http.StatusText(resp.StatusCode))
	}
	// Get the newly put value
	r, err := http.Get(path)
	if err != nil {
		t.Error(err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		t.Fatalf("Expected %q, got %q.", http.StatusText(http.StatusFound),
			http.StatusText(r.StatusCode))
	}
	var body []byte
	switch {
	case strings.Contains(r.Header.Get("Content-Type"), "text/plain"):
		if body, err = io.ReadAll(r.Body); err != nil {
			t.Error(err)
		}
		if !strings.Contains(string(body), val) {
			t.Errorf("Expected %q, got %q.", val,
				string(body))
		}
	default:
		t.Fatalf("Unsupported Content-Type: %q", r.Header.Get("Content-Type"))
	}
	// Delete the value
	req, err = http.NewRequest(
		http.MethodDelete,
		path,
		bytes.NewBuffer([]byte("")),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %q, got %q.", http.StatusText(http.StatusOK),
			http.StatusText(resp.StatusCode))
	}
	// Try and get the value, should fail since it was just deleted
	r, err = http.Get(path)
	if err != nil {
		t.Fatalf("Expected %q, got %q.", http.StatusText(http.StatusNotFound),
			http.StatusText(r.StatusCode))
	}
}
