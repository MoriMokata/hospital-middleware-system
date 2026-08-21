package his

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHospitalAClient_Search_Success(t *testing.T) {
	nationalID := "1234567890123"
	server := NewMockHospitalAServer(map[string]MockHospitalARecord{
		nationalID: {
			FirstNameTH: "สมศรี",
			LastNameTH:  "ใจดี",
			FirstNameEN: "Somsri",
			LastNameEN:  "Jaidee",
			DateOfBirth: "1990-05-12",
			PatientHN:   "HN-000123",
			NationalID:  &nationalID,
			PhoneNumber: "0812345678",
			Email:       "somsri@example.com",
			Gender:      "F",
		},
	})
	defer server.Close()

	client := NewHospitalAClient(server.URL, nil)
	patient, err := client.Search(context.Background(), nationalID)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if patient.FirstNameEN == nil || *patient.FirstNameEN != "Somsri" {
		t.Errorf("FirstNameEN = %v", patient.FirstNameEN)
	}
	if patient.LastName == nil || *patient.LastName != "Jaidee" {
		t.Errorf("LastName (normalized) = %v", patient.LastName)
	}
	if patient.DateOfBirth == nil || patient.DateOfBirth.Format("2006-01-02") != "1990-05-12" {
		t.Errorf("DateOfBirth = %v", patient.DateOfBirth)
	}
	if patient.NationalID == nil || *patient.NationalID != nationalID {
		t.Errorf("NationalID = %v", patient.NationalID)
	}
	if len(patient.RawSource) == 0 {
		t.Error("RawSource should retain the raw HIS response")
	}
}

func TestHospitalAClient_Search_NotFound(t *testing.T) {
	server := NewMockHospitalAServer(map[string]MockHospitalARecord{})
	defer server.Close()

	client := NewHospitalAClient(server.URL, nil)
	_, err := client.Search(context.Background(), "does-not-exist")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Search() error = %v, want ErrNotFound", err)
	}
}

func TestHospitalAClient_Search_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHospitalAClient(server.URL, nil)
	_, err := client.Search(context.Background(), "any-id")
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("Search() error = %v, want a non-nil, non-ErrNotFound error", err)
	}
}

func TestHospitalAClient_Search_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
		case <-r.Context().Done():
		}
	}))
	defer server.Close()

	client := NewHospitalAClient(server.URL, &http.Client{Timeout: 20 * time.Millisecond})
	_, err := client.Search(context.Background(), "any-id")
	if err == nil {
		t.Fatal("Search() expected a timeout error, got nil")
	}
}
