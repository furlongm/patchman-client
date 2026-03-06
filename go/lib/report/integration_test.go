package report

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/furlongm/patchman-client/go/lib/packages"
	"github.com/furlongm/patchman-client/go/lib/repos"
	"github.com/furlongm/patchman-client/go/lib/si"
	"github.com/furlongm/patchman-client/go/lib/updates"
)

func makeTestReport() *Report {
	osInfo := si.OSInfo{
		Hostname: "test-host.example.com",
		OSName:   "Rocky Linux 9.5 (Blue Onyx)",
		OSArch:   "x86_64",
		Kernel:   "5.14.0-503.el9.x86_64",
		CPE:      "cpe:/o:rocky:rocky:9",
	}
	rpt := NewReport(osInfo)
	rpt.SetTags("web,prod")

	rpt.AddRPMs([]packages.RPM{
		{Name: "bash", Epoch: "", Version: "5.1.8", Release: "9.el9", Arch: "x86_64"},
		{Name: "kernel", Epoch: "0", Version: "5.14.0", Release: "503.el9", Arch: "x86_64"},
		{Name: "openssl", Epoch: "1", Version: "3.0.7", Release: "27.el9", Arch: "x86_64"},
	})

	rpt.AddYumRepos([]repos.YumRepo{
		{ID: "baseos", Name: "Rocky Linux 9 - BaseOS", BaseURL: "http://mirror.example.com/rocky/9/BaseOS/x86_64/os/", Priority: 0, Enabled: true},
		{ID: "appstream", Name: "Rocky Linux 9 - AppStream", BaseURL: "http://mirror.example.com/rocky/9/AppStream/x86_64/os/", Priority: 0, Enabled: true},
	}, "cpe:/o:rocky:rocky:9")

	rpt.AddSecurityUpdates([]updates.Update{
		{Name: "openssl", Version: "1:3.0.7-28.el9", Arch: "x86_64", Repo: "baseos"},
	})
	rpt.AddBugfixUpdates([]updates.Update{
		{Name: "bash", Version: "5.1.8-10.el9", Arch: "x86_64", Repo: "baseos"},
	})

	return rpt
}

func TestSubmitToMockServer(t *testing.T) {
	var receivedBody []byte
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/report/" {
			t.Errorf("Expected path /api/report/, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		receivedHeaders = r.Header
		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SubmitResponse{
			Status:  "ok",
			Message: "Report saved successfully",
		})
	}))
	defer server.Close()

	rpt := makeTestReport()
	resp, err := rpt.SubmitWithOptions(SubmitOptions{
		ServerURL:  server.URL,
		VerifySSL:  false,
		MaxRetries: 1,
	})
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}
	if resp.Message != "Report saved successfully" {
		t.Errorf("Expected success message, got '%s'", resp.Message)
	}

	// Verify request content
	if ct := receivedHeaders.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	var parsed Report
	if err := json.Unmarshal(receivedBody, &parsed); err != nil {
		t.Fatalf("Failed to parse submitted JSON: %v", err)
	}
	if parsed.Protocol != 2 {
		t.Errorf("Expected protocol 2, got %d", parsed.Protocol)
	}
	if parsed.Hostname != "test-host.example.com" {
		t.Errorf("Expected hostname test-host.example.com, got %s", parsed.Hostname)
	}
	if len(parsed.Packages) != 3 {
		t.Errorf("Expected 3 packages, got %d", len(parsed.Packages))
	}
	if len(parsed.Repos) != 2 {
		t.Errorf("Expected 2 repos, got %d", len(parsed.Repos))
	}
	if len(parsed.SecUpdates) != 1 {
		t.Errorf("Expected 1 security update, got %d", len(parsed.SecUpdates))
	}
	if len(parsed.BugUpdates) != 1 {
		t.Errorf("Expected 1 bugfix update, got %d", len(parsed.BugUpdates))
	}
	if len(parsed.Tags) != 2 || parsed.Tags[0] != "web" || parsed.Tags[1] != "prod" {
		t.Errorf("Expected tags [web prod], got %v", parsed.Tags)
	}
}

func TestSubmitWithAPIKey(t *testing.T) {
	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SubmitResponse{Status: "ok", Message: "ok"})
	}))
	defer server.Close()

	rpt := makeTestReport()
	_, err := rpt.SubmitWithOptions(SubmitOptions{
		ServerURL:  server.URL,
		APIKey:     "pm_test_key_12345",
		VerifySSL:  false,
		MaxRetries: 1,
	})
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	if receivedAuth != "Api-Key pm_test_key_12345" {
		t.Errorf("Expected 'Api-Key pm_test_key_12345', got '%s'", receivedAuth)
	}
}

func TestSubmitServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SubmitResponse{Error: "internal server error"})
	}))
	defer server.Close()

	rpt := makeTestReport()
	_, err := rpt.SubmitWithOptions(SubmitOptions{
		ServerURL:  server.URL,
		VerifySSL:  false,
		MaxRetries: 1,
	})
	if err == nil {
		t.Fatal("Expected error for 500 response, got nil")
	}
}

func TestSubmitAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(SubmitResponse{Detail: "Authentication credentials were not provided."})
	}))
	defer server.Close()

	rpt := makeTestReport()
	_, err := rpt.SubmitWithOptions(SubmitOptions{
		ServerURL:  server.URL,
		VerifySSL:  false,
		MaxRetries: 3, // Should NOT retry on 4xx
	})
	if err == nil {
		t.Fatal("Expected error for 403 response, got nil")
	}
}

func TestSubmitRetryOnServerError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(SubmitResponse{Error: "bad gateway"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SubmitResponse{Status: "ok", Message: "ok"})
	}))
	defer server.Close()

	rpt := makeTestReport()
	resp, err := rpt.SubmitWithOptions(SubmitOptions{
		ServerURL:  server.URL,
		VerifySSL:  false,
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Expected success after retries, got: %v", err)
	}
	if resp.Message != "ok" {
		t.Errorf("Expected 'ok', got '%s'", resp.Message)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestSubmitEmptyReport(t *testing.T) {
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SubmitResponse{Status: "ok", Message: "ok"})
	}))
	defer server.Close()

	osInfo := si.OSInfo{Hostname: "empty-host", OSArch: "x86_64", Kernel: "5.10.0"}
	rpt := NewReport(osInfo)
	_, err := rpt.SubmitWithOptions(SubmitOptions{
		ServerURL:  server.URL,
		VerifySSL:  false,
		MaxRetries: 1,
	})
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	var parsed Report
	if err := json.Unmarshal(receivedBody, &parsed); err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	if len(parsed.Packages) != 0 {
		t.Errorf("Expected 0 packages, got %d", len(parsed.Packages))
	}
	if len(parsed.Repos) != 0 {
		t.Errorf("Expected 0 repos, got %d", len(parsed.Repos))
	}
}
