package api

import (
	"net/http"

	"backend/internal/generated/api/generated"
	"backend/internal/response"

	"go.uber.org/zap"
)

// ListClub100People returns all 100 Club members with their redemption status.
// (GET /club100/people)
func (h *Handlers) ListClub100People(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	people, err := h.club100.GetPeopleWithRedemptions(ctx)
	if err != nil {
		h.logger.Error("club100 GetPeopleWithRedemptions failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "club100_error", err.Error())
		return
	}

	items := make([]generated.Club100Person, 0, len(people))
	for _, p := range people {
		items = append(items, generated.Club100Person{
			Id:        p.ID,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Remaining: p.Remaining,
			Max:       p.Max,
		})
	}
	response.WriteJSON(w, http.StatusOK, generated.Club100PersonList{Items: items})
}

// GetClub100Remaining returns the remaining redemptions for a specific person.
// (GET /club100/remaining/{elvantoPersonId})
func (h *Handlers) GetClub100Remaining(w http.ResponseWriter, r *http.Request, elvantoPersonId string) {
	ctx := r.Context()

	remaining, max, err := h.club100.GetRemainingRedemptions(ctx, elvantoPersonId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "club100_error", err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, generated.Club100Remaining{
		ElvantoPersonId: elvantoPersonId,
		Remaining:       remaining,
		Max:             max,
	})
}
