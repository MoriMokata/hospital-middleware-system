package his

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockHospitalARecord is the seed data for NewMockHospitalAServer, keyed by
// the id (national_id or passport_id) a client would search for.
type MockHospitalARecord = hospitalAResponse

// NewMockHospitalAServer starts an httptest server that mimics Hospital A's
// GET /patient/search/{id} endpoint (docs/api-spec.md), for use in tests
// and local dev in place of the real hospital-a.api.co.th. The caller must
// Close() the returned server.
func NewMockHospitalAServer(records map[string]MockHospitalARecord) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/patient/search/")
		record, ok := records[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	}))
}
