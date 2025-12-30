package iid

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestBatchManage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(map[string]any{"results": []map[string]any{{}, {}}})
	}))
	defer srv.Close()

	old := IIDBaseURL
	IIDBaseURL = srv.URL + "/iid/v1"
	defer func() { IIDBaseURL = old }()

	oldEnv := os.Getenv("PUSHGW_TEST_NO_OAUTH")
	_ = os.Setenv("PUSHGW_TEST_NO_OAUTH", "1")
	defer os.Setenv("PUSHGW_TEST_NO_OAUTH", oldEnv)

	client, err := NewFromCreds(context.Background(), []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.BatchManage("movies", []string{"t1", "t2"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
}

func TestBatchManage_PartialFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"results": []map[string]any{{"error": "NOT_FOUND"}, {}}})
	}))
	defer srv.Close()

	old := IIDBaseURL
	IIDBaseURL = srv.URL + "/iid/v1"
	defer func() { IIDBaseURL = old }()

	oldEnv := os.Getenv("PUSHGW_TEST_NO_OAUTH")
	_ = os.Setenv("PUSHGW_TEST_NO_OAUTH", "1")
	defer os.Setenv("PUSHGW_TEST_NO_OAUTH", oldEnv)

	client, err := NewFromCreds(context.Background(), []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.BatchManage("movies", []string{"t1", "t2"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
	if _, ok := res[0]["error"]; !ok {
		t.Fatalf("expected error in first result, got %#v", res[0])
	}
}

func TestNewFromCreds_Missing(t *testing.T) {
	_, err := NewFromCreds(context.Background(), nil)
	if err == nil {
		t.Fatalf("expected error when credentials are missing")
	}
}

func TestBatchManage_HeaderAndResults(t *testing.T) {
	// Server checks the access_token_auth header and echoes results
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("access_token_auth") != "true" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(map[string]any{"results": []map[string]any{{}, {}}})
	}))
	defer srv.Close()

	old := IIDBaseURL
	IIDBaseURL = srv.URL + "/iid/v1"
	defer func() { IIDBaseURL = old }()

	oldEnv := os.Getenv("PUSHGW_TEST_NO_OAUTH")
	_ = os.Setenv("PUSHGW_TEST_NO_OAUTH", "1")
	defer os.Setenv("PUSHGW_TEST_NO_OAUTH", oldEnv)

	client, err := NewFromCreds(context.Background(), []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.BatchManage("movies", []string{"t1", "t2"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
}

func TestBatchManage_NonJSONBody(t *testing.T) {
	// Server returns plain text; BatchManage should not crash and should return empty parsed.Results
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	old := IIDBaseURL
	IIDBaseURL = srv.URL + "/iid/v1"
	defer func() { IIDBaseURL = old }()

	oldEnv := os.Getenv("PUSHGW_TEST_NO_OAUTH")
	_ = os.Setenv("PUSHGW_TEST_NO_OAUTH", "1")
	defer os.Setenv("PUSHGW_TEST_NO_OAUTH", oldEnv)

	client, err := NewFromCreds(context.Background(), []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.BatchManage("movies", []string{"t1", "t2"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Fatalf("expected 0 results for invalid JSON, got %d", len(res))
	}
}

func TestBatchManage_InvalidCredsWithoutNoOAuth(t *testing.T) {
	// Ensure when PUSHGW_TEST_NO_OAUTH not set, invalid creds lead to an error from BatchManage
	oldEnv := os.Getenv("PUSHGW_TEST_NO_OAUTH")
	_ = os.Unsetenv("PUSHGW_TEST_NO_OAUTH")
	defer os.Setenv("PUSHGW_TEST_NO_OAUTH", oldEnv)

	client, err := NewFromCreds(context.Background(), []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}

	// Use a fast deadline so a broken token exchange doesn't hang the test.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c := &IIDClient{ctx: ctx, creds: client.creds}
	_, err = c.BatchManage("movies", []string{"t1"}, true)
	if err == nil {
		t.Fatalf("expected error for invalid creds when not skipping OAuth")
	}
}

func TestConvertToTNPGResponses_Mapping(t *testing.T) {
	parsed := []map[string]any{{"error": "NOT_FOUND"}, {}}
	tokens := []string{"t1", "t2", "t3"}
	res := ConvertToTNPGResponses(parsed, tokens)
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if len(res.Responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(res.Responses))
	}
	if res.FailureCount != 1 {
		t.Fatalf("expected FailureCount 1, got %d", res.FailureCount)
	}
	if res.SuccessCount != 2 {
		t.Fatalf("expected SuccessCount 2, got %d", res.SuccessCount)
	}
	if res.Responses[0].ErrorCode != "NOT_FOUND" {
		t.Fatalf("expected error code NOT_FOUND for first token, got %q", res.Responses[0].ErrorCode)
	}
}
