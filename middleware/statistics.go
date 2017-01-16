package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

type StatisticsMiddleware struct {
	mutex             sync.RWMutex
	totalRequests     int
	numRecentRequests int
	recentRequests    []RequestsRecord
}

type RequestsRecord struct {
	Method     string        `json:"method"`
	Host       string        `json:"host"`
	Path       string        `json:"path"`
	StatusCode int           `json:"status_code"`
	Status     string        `json:"status"`
	Delay      time.Duration `json:"delay"`
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.ResponseWriter.WriteHeader(statusCode)
	rr.statusCode = statusCode
}

func NewStatisticsMiddleware(numRecentRequests int) *StatisticsMiddleware {
	return &StatisticsMiddleware{
		numRecentRequests: numRecentRequests,
		recentRequests:    []RequestsRecord{},
	}
}

func (m *StatisticsMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	respRecorder := &responseRecorder{rw, http.StatusOK}
	begin := time.Now()
	next(respRecorder, r)
	if strings.HasPrefix(r.URL.Path, "/states") || strings.HasPrefix(r.URL.Path, "/statistics") {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.totalRequests++
	m.recentRequests = append(m.recentRequests, RequestsRecord{
		StatusCode: respRecorder.statusCode,
		Status:     http.StatusText(respRecorder.statusCode),
		Method:     r.Method,
		Host:       r.Host,
		Path:       r.URL.String(),
		Delay:      time.Now().Sub(begin),
	})
	if len(m.recentRequests) > m.numRecentRequests {
		m.recentRequests = m.recentRequests[len(m.recentRequests)-m.numRecentRequests:]
	}
}

func (m *StatisticsMiddleware) Json() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	statisticsMap := make(map[string]interface{})
	statisticsMap["totalRequests"] = m.totalRequests
	statisticsMap["recentRequests"] = m.recentRequests

	buf, _ := json.Marshal(statisticsMap)
	return string(buf)
}
