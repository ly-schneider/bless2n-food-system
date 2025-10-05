package handler

import (
	_ "backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/service"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandler struct {
	userService service.UserService
	validator   *validator.Validate
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator.New(),
	}
}

// GetCurrent returns the authenticated user's full profile data.
// GET /v1/users (auth required)
func (h *UserHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok || claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	oid, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid user id"))
		return
	}
	u, err := h.userService.GetByID(r.Context(), oid)
	if err != nil || u == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user": map[string]any{
			"id":         u.ID.Hex(),
			"email":      u.Email,
			"role":       u.Role,
			"firstName":  u.FirstName,
			"lastName":   u.LastName,
			"isVerified": u.IsVerified,
			"createdAt":  u.CreatedAt,
			"updatedAt":  u.UpdatedAt,
		},
	})
}

type requestEmailChangeBody struct {
	NewEmail string `json:"newEmail" validate:"required,email"`
}

// RequestEmailChange starts the email change flow by sending a code to the new email
func (h *UserHandler) RequestEmailChange(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body requestEmailChangeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid JSON payload"))
		return
	}
	if err := h.validator.Struct(body); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(verrs), r.URL.Path))
			return
		}
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Validation failed"))
		return
	}
	oid, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid user id"))
		return
	}
	if err := h.userService.RequestEmailChange(r.Context(), oid, body.NewEmail, clientIP(r), clientIDFromRequest(r)); err != nil {
		// map errors to HTTP
		switch err.Error() {
		case "same_email":
			response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "New email matches current email"))
		case "email_taken":
			response.WriteProblem(w, response.NewProblem(http.StatusConflict, http.StatusText(http.StatusConflict), "Email already in use"))
		case "invalid_email":
			response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid email"))
		default:
			http.Error(w, "Failed to initiate change", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "Verification code sent to new email"})
}

type confirmEmailChangeBody struct {
	Code string `json:"code" validate:"required,len=6"`
}

// ConfirmEmailChange verifies the code and updates the user's email across systems
func (h *UserHandler) ConfirmEmailChange(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body confirmEmailChangeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid JSON payload"))
		return
	}
	if err := h.validator.Struct(body); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(verrs), r.URL.Path))
			return
		}
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Validation failed"))
		return
	}
	oid, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid user id"))
		return
	}
	u, err := h.userService.ConfirmEmailChange(r.Context(), oid, body.Code)
	if err != nil {
		switch err.Error() {
		case "invalid_code":
			http.Error(w, "Invalid code", http.StatusUnauthorized)
		case "too_many_attempts":
			http.Error(w, "Too many attempts", http.StatusTooManyRequests)
		case "email_taken":
			response.WriteProblem(w, response.NewProblem(http.StatusConflict, http.StatusText(http.StatusConflict), "Email already in use"))
		default:
			http.Error(w, "Failed to confirm", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"user": map[string]any{"id": u.ID.Hex(), "email": u.Email, "role": u.Role}})
}

// DeleteUser deletes the authenticated user's account and related auth artifacts.
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	oid, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid user id"))
		return
	}
	if err := h.userService.DeleteAccount(r.Context(), oid); err != nil {
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}
	// 204 No Content per REST conventions
	w.WriteHeader(http.StatusNoContent)
}

// UpdateUser allows the authenticated user to update allowed fields. If email changes, a code is sent to confirm.
type updateUserBody struct {
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body updateUserBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid JSON payload"))
		return
	}
	if err := h.validator.Struct(body); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(verrs), r.URL.Path))
			return
		}
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Validation failed"))
		return
	}
	oid, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid user id"))
		return
	}
	user, emailInit, err := h.userService.UpdateProfile(r.Context(), oid, body.FirstName, body.LastName, body.Email, claims.Role, clientIP(r), clientIDFromRequest(r))
	if err != nil {
		switch err.Error() {
		case "email_taken":
			response.WriteProblem(w, response.NewProblem(http.StatusConflict, http.StatusText(http.StatusConflict), "Email already in use"))
		case "invalid_email":
			response.WriteProblem(w, response.NewProblem(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid email"))
		case "same_email":
			// nothing to do; still return the current user
		default:
			http.Error(w, "Failed to update", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"user": map[string]any{"id": user.ID.Hex(), "email": user.Email, "role": user.Role, "firstName": user.FirstName, "lastName": user.LastName},
	}
	if emailInit {
		resp["email_change_initiated"] = true
		resp["message"] = "Verification code sent to new email"
	}
	_ = json.NewEncoder(w).Encode(resp)
}
