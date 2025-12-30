package iid

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/tinode/pushtype"
	"golang.org/x/oauth2/google"
)

// Default IID base URL, overridable in tests.
var IIDBaseURL = "https://iid.googleapis.com/iid/v1"

// IIDClient handles Instance ID batch operations using a service account JSON.
type IIDClient struct {
	ctx   context.Context
	creds []byte
}

// NewFromCreds creates a client from raw service account JSON.
func NewFromCreds(ctx context.Context, creds []byte) (*IIDClient, error) {
	if creds == nil {
		return nil, errors.New("missing credentials for IID calls")
	}
	return &IIDClient{ctx: ctx, creds: creds}, nil
}

// BatchManage performs add/remove of registration tokens to/from a topic.
// Returns per-token results as parsed from IID API: []map[string]any (results field).
func (c *IIDClient) BatchManage(topic string, tokens []string, add bool) ([]map[string]any, error) {
	if len(tokens) == 0 {
		return []map[string]any{}, nil
	}
	if c.creds == nil {
		return nil, errors.New("missing credentials for IID calls")
	}

	// Create HTTP client. For tests we can skip OAuth token exchange by setting
	// PUSHGW_TEST_NO_OAUTH=1 which uses plain http.Client (the mock server will accept
	// requests regardless of Authorization header).
	var client *http.Client
	if os.Getenv("PUSHGW_TEST_NO_OAUTH") == "1" {
		client = &http.Client{}
	} else {
		conf, err := google.JWTConfigFromJSON(c.creds, "https://www.googleapis.com/auth/firebase.messaging")
		if err != nil {
			return nil, err
		}
		client = conf.Client(c.ctx)
	}

	var url string
	if add {
		url = IIDBaseURL + ":batchAdd"
	} else {
		url = IIDBaseURL + ":batchRemove"
	}

	body := map[string]any{"to": "/topics/" + topic, "registration_tokens": tokens}
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// Per docs, include this header so server knows we're using access token auth.
	req.Header.Set("access_token_auth", "true")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var parsed struct {
		Results []map[string]any `json:"results"`
	}
	_ = json.Unmarshal(respBody, &parsed)

	return parsed.Results, nil
}

// ConvertToTNPGResponses converts parsed IID results to an array of TNPGResponse matched to tokens.
func ConvertToTNPGResponses(parsed []map[string]any, tokens []string) *pushtype.BatchResponse {
	out := &pushtype.BatchResponse{}
	out.Responses = make([]*pushtype.TNPGResponse, len(tokens))
	for i := range tokens {
		if i < len(parsed) {
			if errStr, ok := parsed[i]["error"].(string); ok && errStr != "" {
				out.Responses[i] = &pushtype.TNPGResponse{ErrorCode: errStr, ErrorMessage: errStr}
				out.FailureCount++
			} else {
				out.Responses[i] = &pushtype.TNPGResponse{Code: 200}
				out.SuccessCount++
			}
		} else {
			// No per-item result: treat as success.
			out.Responses[i] = &pushtype.TNPGResponse{Code: 200}
			out.SuccessCount++
		}
	}
	return out
}
