package deleteurl

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type DeleteURL interface {
	Delete(alias string) error
}

func New(log *slog.Logger, del DeleteURL) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const op = "handlers/url/gelete/New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(req.Context())),
		)

		alias := chi.URLParam(req, "alias")
		if alias == "" {
			log.Info("alias is empty")
			w.WriteHeader(http.StatusBadRequest)
			log.Error("alias is required")
			return
		}
		err := del.Delete(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", slog.String("alias", alias))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Info("url deleted", slog.String("alias", alias))
		w.WriteHeader(http.StatusNoContent)
	}
}
