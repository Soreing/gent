package gent

import (
	"net/http"
	"net/http/httptest"
	"time"
)

type mockRequester struct {
	// Request
	LastRequest *http.Request
	CountCalled int

	// Response
	Delay      time.Duration
	RequestErr error
	StatusCode int

	// Closed
	ClosedCount int
}

func (m *mockRequester) CloseIdleConnections() {
	m.ClosedCount++
}

func (m *mockRequester) Do(r *http.Request) (*http.Response, error) {
	m.CountCalled++
	m.LastRequest = r

	time.Sleep(m.Delay)

	if m.RequestErr != nil {
		return nil, m.RequestErr
	} else {
		rec := httptest.NewRecorder()
		res := rec.Result()
		res.StatusCode = m.StatusCode
		return res, nil
	}
}
