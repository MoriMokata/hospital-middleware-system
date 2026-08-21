package his

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
)

// HospitalAClient implements HISClient against Hospital A's
// GET /patient/search/{id} endpoint (see docs/api-spec.md).
type HospitalAClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewHospitalAClient(baseURL string, httpClient *http.Client) *HospitalAClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &HospitalAClient{BaseURL: strings.TrimRight(baseURL, "/"), HTTPClient: httpClient}
}

// hospitalAResponse mirrors the JSON body documented in
// docs/api-spec.md#internaldev-only-hospital-a-mock.
type hospitalAResponse struct {
	FirstNameTH  string  `json:"first_name_th"`
	MiddleNameTH *string `json:"middle_name_th"`
	LastNameTH   string  `json:"last_name_th"`
	FirstNameEN  string  `json:"first_name_en"`
	MiddleNameEN *string `json:"middle_name_en"`
	LastNameEN   string  `json:"last_name_en"`
	DateOfBirth  string  `json:"date_of_birth"`
	PatientHN    string  `json:"patient_hn"`
	NationalID   *string `json:"national_id"`
	PassportID   *string `json:"passport_id"`
	PhoneNumber  string  `json:"phone_number"`
	Email        string  `json:"email"`
	Gender       string  `json:"gender"`
}

func (c *HospitalAClient) Search(ctx context.Context, id string) (domain.Patient, error) {
	endpoint := fmt.Sprintf("%s/patient/search/%s", c.BaseURL, url.PathEscape(id))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.Patient{}, fmt.Errorf("hospital a: build request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return domain.Patient{}, fmt.Errorf("hospital a: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.Patient{}, fmt.Errorf("hospital a: read response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var payload hospitalAResponse
		if err := json.Unmarshal(body, &payload); err != nil {
			return domain.Patient{}, fmt.Errorf("hospital a: decode response: %w", err)
		}
		return mapHospitalAResponse(payload, body), nil
	case http.StatusNotFound:
		return domain.Patient{}, ErrNotFound
	default:
		return domain.Patient{}, fmt.Errorf("hospital a: unexpected status %d", resp.StatusCode)
	}
}

func mapHospitalAResponse(r hospitalAResponse, raw []byte) domain.Patient {
	p := domain.Patient{
		FirstNameTH:  strPtr(r.FirstNameTH),
		MiddleNameTH: r.MiddleNameTH,
		LastNameTH:   strPtr(r.LastNameTH),
		FirstNameEN:  strPtr(r.FirstNameEN),
		MiddleNameEN: r.MiddleNameEN,
		LastNameEN:   strPtr(r.LastNameEN),
		FirstName:    strPtr(r.FirstNameEN),
		MiddleName:   r.MiddleNameEN,
		LastName:     strPtr(r.LastNameEN),
		PatientHN:    strPtr(r.PatientHN),
		NationalID:   r.NationalID,
		PassportID:   r.PassportID,
		PhoneNumber:  strPtr(r.PhoneNumber),
		Email:        strPtr(r.Email),
		Gender:       strPtr(r.Gender),
		RawSource:    json.RawMessage(raw),
	}
	if dob, err := time.Parse("2006-01-02", r.DateOfBirth); err == nil {
		p.DateOfBirth = &dob
	}
	return p
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
