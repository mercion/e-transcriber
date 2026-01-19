package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleSessionMethodNotAllowed(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/session", nil)
	w := httptest.NewRecorder()

	srv.handleSession(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleSessionMissingSDP(t *testing.T) {
	srv := &Server{}
	body := bytes.NewBufferString(`{"type":"offer"}`)
	req := httptest.NewRequest(http.MethodPost, "/session", body)
	w := httptest.NewRecorder()

	srv.handleSession(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
