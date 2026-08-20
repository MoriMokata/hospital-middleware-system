package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MoriMokata/hospital-middleware-system/internal/middleware"
	"github.com/MoriMokata/hospital-middleware-system/internal/pkg"
	"github.com/MoriMokata/hospital-middleware-system/internal/service"
)

// PatientHandler wires /patient/search to PatientService. Must run behind
// middleware.Auth so hospital_id is present in the request context.
type PatientHandler struct {
	Patients *service.PatientService
}

func NewPatientHandler(patients *service.PatientService) *PatientHandler {
	return &PatientHandler{Patients: patients}
}

type searchPatientRequest struct {
	NationalID  *string `json:"national_id"`
	PassportID  *string `json:"passport_id"`
	FirstName   *string `json:"first_name"`
	MiddleName  *string `json:"middle_name"`
	LastName    *string `json:"last_name"`
	DateOfBirth *string `json:"date_of_birth"`
	PhoneNumber *string `json:"phone_number"`
	Email       *string `json:"email"`
}

type patientResponse struct {
	ID           string  `json:"id"`
	PatientHN    *string `json:"patient_hn"`
	NationalID   *string `json:"national_id"`
	PassportID   *string `json:"passport_id"`
	FirstNameTH  *string `json:"first_name_th"`
	MiddleNameTH *string `json:"middle_name_th"`
	LastNameTH   *string `json:"last_name_th"`
	FirstNameEN  *string `json:"first_name_en"`
	MiddleNameEN *string `json:"middle_name_en"`
	LastNameEN   *string `json:"last_name_en"`
	DateOfBirth  *string `json:"date_of_birth"`
	PhoneNumber  *string `json:"phone_number"`
	Email        *string `json:"email"`
	Gender       *string `json:"gender"`
}

// Search handles POST /patient/search (docs/api-spec.md#post-patientsearch).
func (h *PatientHandler) Search(c *gin.Context) {
	hospitalID, ok := middleware.HospitalIDFromContext(c)
	if !ok || hospitalID == "" {
		c.JSON(http.StatusUnauthorized, pkg.NewErrorEnvelope("UNAUTHORIZED", "missing or invalid token"))
		return
	}

	var req searchPatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewErrorEnvelope("VALIDATION_ERROR", "invalid request body"))
		return
	}

	results, err := h.Patients.Search(c.Request.Context(), hospitalID, service.SearchInput{
		NationalID:  req.NationalID,
		PassportID:  req.PassportID,
		FirstName:   req.FirstName,
		MiddleName:  req.MiddleName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
	})
	if err != nil {
		var verr *service.ValidationError
		if errors.As(err, &verr) {
			c.JSON(http.StatusBadRequest, pkg.NewErrorEnvelope("VALIDATION_ERROR", verr.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, pkg.NewErrorEnvelope("INTERNAL_ERROR", "unexpected error"))
		return
	}

	body := make([]patientResponse, len(results))
	for i, p := range results {
		body[i] = patientResponse{
			ID:           p.ID.String(),
			PatientHN:    p.PatientHN,
			NationalID:   p.NationalID,
			PassportID:   p.PassportID,
			FirstNameTH:  p.FirstNameTH,
			MiddleNameTH: p.MiddleNameTH,
			LastNameTH:   p.LastNameTH,
			FirstNameEN:  p.FirstNameEN,
			MiddleNameEN: p.MiddleNameEN,
			LastNameEN:   p.LastNameEN,
			DateOfBirth:  formatDate(p.DateOfBirth),
			PhoneNumber:  p.PhoneNumber,
			Email:        p.Email,
			Gender:       p.Gender,
		}
	}

	c.JSON(http.StatusOK, gin.H{"results": body, "count": len(body)})
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}
