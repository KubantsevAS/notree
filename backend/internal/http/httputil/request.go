package httputil

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func PathParamIDAndUser(w http.ResponseWriter, r *http.Request) (parsedID, userID pgtype.UUID, ok bool) {
	userID, err := GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		WriteErrorJSON(w, "unauthorized", http.StatusUnauthorized)
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	pathID := chi.URLParam(r, "id")
	parsedID, err = PgUUIDFromString(&pathID)
	if err != nil {
		WriteErrorJSON(w, "invalid node id format", http.StatusBadRequest)
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	return parsedID, userID, true
}
