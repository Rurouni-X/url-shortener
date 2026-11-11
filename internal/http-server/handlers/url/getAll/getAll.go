package getall

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/database"

	"github.com/go-chi/chi/v5/middleware"
)

type GetURL interface {
	GetAll() ([]database.Results, error)
}


func New(log *slog.Logger, GetUrl GetURL) http.HandlerFunc {
	
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers/url/getAll/New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		ressult, err := GetUrl.GetAll()
		if err != nil {
			log.Error("failed to get all urls", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		
		if err := json.NewEncoder(w).Encode(ressult); err != nil {
			log.Error("failed to encode request body", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}