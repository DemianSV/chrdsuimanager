package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func getDataBaseStatus(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			responseJSON, err := json.Marshal(DBStatusA)
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/database/status", "500", "GET", "getDataBaseStatus").Set(duration)

				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/database/status", "200", "GET", "getDataBaseStatus").Set(duration)

				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
				return
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/database/status", "500", "GET", "getDataBaseStatus").Set(duration)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/database/status", "403", "GET", "getDataBaseStatus").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}
