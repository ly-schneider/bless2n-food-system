package api

import (
	"encoding/json"
	"net/http"

	"backend/internal/generated/api/generated"
	"backend/internal/response"
)

// ListUsers returns users, optionally filtered by role.
// (GET /users)
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request, params generated.ListUsersParams) {
	ctx := r.Context()

	var roleFilter *string
	if params.Role != nil {
		s := string(*params.Role)
		roleFilter = &s
	}

	// After Task 6, a UserService will expose List/Get/Update/Delete
	// that return ent.User types. For now this calls the future interface.
	users, _, err := h.users.List(ctx, roleFilter)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, generated.UserList{
		Items: toAPIUsers(users),
	})
}

// GetUser returns a single user by their Better Auth user ID (string).
// (GET /users/{userId})
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request, userId string) {
	user, err := h.users.GetByID(r.Context(), userId)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIUser(user))
}

// UpdateUser partially updates a user.
// (PATCH /users/{userId})
func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request, userId string) {
	var body generated.UserUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	ctx := r.Context()

	if body.Role != nil {
		if err := h.users.UpdateRole(ctx, userId, string(*body.Role)); err != nil {
			writeEntError(w, err)
			return
		}
	}

	if body.Name != nil {
		if err := h.users.UpdateName(ctx, userId, *body.Name); err != nil {
			writeEntError(w, err)
			return
		}
	}

	user, err := h.users.GetByID(ctx, userId)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIUser(user))
}

// DeleteUser removes a user.
// (DELETE /users/{userId})
func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request, userId string) {
	if err := h.users.Delete(r.Context(), userId); err != nil {
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
