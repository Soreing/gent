package gent

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"
)

type mockMemPool struct {
	MemoryPool
}

type mockHttpClient struct {
	HttpClient
}

type mockRetrier struct {
	retrier
}

type mockHttpHandler struct {
	dur  time.Duration
	err  error
	code int

	method   string
	data     []byte
	endpoint []byte
	headers  map[string]string
	called   int
}

func (m *mockHttpHandler) Do(r *http.Request) (*http.Response, error) {
	m.called++
	m.method = r.Method
	m.data, _ = ioutil.ReadAll(r.Body)
	m.endpoint = []byte(r.URL.String())

	for k, v := range r.Header {
		if len(v) > 0 {
			m.headers[k] = v[0]
		}
	}

	select {
	case <-r.Context().Done():
		return nil, r.Context().Err()
	case <-time.NewTimer(m.dur).C:
		if m.err != nil {
			return nil, m.err
		} else {
			rec := httptest.NewRecorder()
			res := rec.Result()
			res.StatusCode = m.code
			return res, nil
		}
	}
}
