package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/DemianSV/chrdsclient"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func getDataManagerStatus(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	type ResponseT struct {
		Up bool `json:"up"`
	}
	type ResponseAT []ResponseT

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {

			var response ResponseT
			var responseA ResponseAT

			status := chrdsclient.Status()

			for _, item := range status {
				response.Up = item
				responseA = append(responseA, response)
			}

			responseJSON, err := json.Marshal(responseA)
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/datamanager/status", "500", "GET", "getDataManagerStatus").Set(duration)

				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/datamanager/status", "200", "GET", "getDataManagerStatus").Set(duration)

				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
				return
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/datamanager/status", "500", "GET", "getDataManagerStatus").Set(duration)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/datamanager/status", "403", "GET", "getDataManagerStatus").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}
