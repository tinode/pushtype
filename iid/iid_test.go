package iid

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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
