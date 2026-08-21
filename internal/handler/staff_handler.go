package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/MoriMokata/hospital-middleware-system/internal/pkg"
	"github.com/MoriMokata/hospital-middleware-system/internal/service"
)

// StaffHandler wires the /staff/* endpoints to StaffService. Handlers stay
// thin: parse/validate the request shape, call the service, map the result
// to an HTTP response.
type StaffHandler struct {
	Staff *service.StaffService
}

func NewStaffHandler(staff *service.StaffService) *StaffHandler {
	return &StaffHandler{Staff: staff}
}

type createStaffRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hospital string `json:"hospital"`
}

// Create handles POST /staff/create (docs/api-spec.md#post-staffcreate).
func (h *StaffHandler) Create(c *gin.Context) {
	var req createStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewErrorEnvelope("VALIDATION_ERROR", "invalid request body"))
		return
	}

	out, err := h.Staff.CreateStaff(c.Request.Context(), service.CreateStaffInput{
		Username: req.Username,
		Password: req.Password,
		Hospital: req.Hospital,
	})
	if err != nil {
		writeStaffError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         out.ID,
		"username":   out.Username,
		"hospital":   out.Hospital,
		"created_at": out.CreatedAt,
	})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hospital string `json:"hospital"`
}

// Login handles POST /staff/login (docs/api-spec.md#post-stafflogin).
func (h *StaffHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewErrorEnvelope("VALIDATION_ERROR", "invalid request body"))
		return
	}

	out, err := h.Staff.Login(c.Request.Context(), service.LoginInput{
		Username: req.Username,
		Password: req.Password,
		Hospital: req.Hospital,
	})
	if err != nil {
		writeStaffError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": out.AccessToken,
		"token_type":   "Bearer",
		"expires_in":   out.ExpiresIn,
	})
}

func writeStaffError(c *gin.Context, err error) {
	var verr *service.ValidationError
	switch {
	case errors.As(err, &verr):
		c.JSON(http.StatusBadRequest, pkg.NewErrorEnvelope("VALIDATION_ERROR", verr.Error()))
	case errors.Is(err, service.ErrHospitalNotFound):
		c.JSON(http.StatusNotFound, pkg.NewErrorEnvelope("HOSPITAL_NOT_FOUND", "hospital not found"))
	case errors.Is(err, service.ErrUsernameTaken):
		c.JSON(http.StatusConflict, pkg.NewErrorEnvelope("USERNAME_TAKEN", "username already exists for this hospital"))
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, pkg.NewErrorEnvelope("INVALID_CREDENTIALS", "username or password is incorrect"))
	default:
		c.JSON(http.StatusInternalServerError, pkg.NewErrorEnvelope("INTERNAL_ERROR", "unexpected error"))
	}
}
