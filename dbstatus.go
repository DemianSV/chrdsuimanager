package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func getDataBaseStatus(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			responseJSON, err := json.Marshal(DBStatusA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
				return
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}
