package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gocql/gocql"
)

type ResponsModuleT struct {
	ID          string `json:"id"`
	TypeMod     string `json:"type"`
	Status      int    `json:"status"`
	Description string `json:"description"`
	UserID      string `json:"userid"`
}
type ResponseModuleAT []ResponsModuleT

func getModuleCount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			switch claims["role"] {
			case "superadmin":
				scanner = Session.Query(`SELECT count(*) FROM registration`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			case "admin":
				scanner = Session.Query(`SELECT count(*) FROM user_registration WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			default:
				scanner = Session.Query(`SELECT count(*) FROM user_registration WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			}
			for scanner.Next() {
				var count int = 0
				err := scanner.Scan(&count)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseCount.Count = count
					responseCountJSON, err := json.Marshal(responseCount)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseCountJSON)
						return
					}
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
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

func getModuleSelect(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseRegistration ResponsModuleT
			var responseRegistrationA ResponseModuleAT

			ctx := context.Background()
			var scanner gocql.Scanner

			switch claims["role"] {
			case "superadmin":
				scanner = Session.Query(`SELECT id, type, status, description, user_id FROM registration`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			case "admin":
				var registrationList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						registrationList = append(registrationList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range registrationList {
					args = append(args, item)
				}
				if len(registrationList) > 0 {
					scanner = Session.Query(`SELECT id, type, status, description, user_id FROM registration WHERE id IN (?`+strings.Repeat(", ?", len(registrationList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseRegistrationJSON, err := json.Marshal(responseRegistrationA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseRegistrationJSON)
						return
					}
				}
			default:
				var registrationList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						registrationList = append(registrationList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range registrationList {
					args = append(args, item)
				}
				if len(registrationList) > 0 {
					scanner = Session.Query(`SELECT id, type, status, description, user_id FROM registration WHERE id IN (?`+strings.Repeat(", ?", len(registrationList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseRegistrationJSON, err := json.Marshal(responseRegistrationA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseRegistrationJSON)
						return
					}
				}
			}
			for scanner.Next() {
				var id string
				var description string
				var typeMod string
				var status int
				var userID string

				err := scanner.Scan(&id, &typeMod, &status, &description, &userID)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseRegistration.ID = id
					responseRegistration.TypeMod = typeMod
					responseRegistration.Status = status
					responseRegistration.Description = description
					responseRegistration.UserID = userID

					responseRegistrationA = append(responseRegistrationA, responseRegistration)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseRegistrationJSON, err := json.Marshal(responseRegistrationA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseRegistrationJSON)
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

func putModuleUpdate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the module update!", claims["username"])

	type Registration struct {
		ID          string `json:"id"`
		TypeMod     string `json:"type"`
		Status      int    `json:"status"`
		Description string `json:"description"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var registration Registration
			log.Print(string(b))
			if err := json.Unmarshal(b, &registration); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`UPDATE registration SET type = ?, status = ?, description = ? WHERE id = ?`, registration.TypeMod, registration.Status, registration.Description, registration.ID).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to change the module data (UPDATE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully changed the module data %v!", claims["username"], registration.ID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to change the data!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putModuleCreate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of the module!", claims["username"])

	type Registration struct {
		ID          string `json:"id"`
		TypeMod     string `json:"type"`
		Status      int    `json:"status"`
		Description string `json:"description"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			id, err := gocql.RandomUUID()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var registration Registration
			if err := json.Unmarshal(b, &registration); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`
				BEGIN BATCH
					INSERT INTO registration (id, type, status, description, user_id) VALUES (?, ?, ?, ?, ?)
					INSERT INTO user_registration (user_id, registration_id) VALUES (?, ?)
				APPLY BATCH
				`, id, registration.TypeMod, registration.Status, registration.Description, claims["userid"], claims["userid"], id).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to create a module (BATCH: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully created a module!", claims["username"])
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to create a module!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putModuleRemove(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the removal of the module!", claims["username"])

	type Registration struct {
		ID string `json:"id"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var registration Registration
			if err := json.Unmarshal(b, &registration); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()

				var userID string
				if err := Session.Query(`SELECT user_id FROM registration WHERE id = ? LIMIT 1`, registration.ID).WithContext(ctx).Consistency(ConsistencyRead).Scan(&userID); err != nil {
					log.Print("Failed to remove the module (SELECT user_id: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				err := Session.Query(`
				BEGIN BATCH
					DELETE FROM registration WHERE id = ?
					DELETE FROM user_registration WHERE user_id = ? AND registration_id = ?
				APPLY BATCH
			`, registration.ID, userID, registration.ID).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to remove the module (DELETE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully deleted the module %v!", claims["username"], registration.ID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to remove the module!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}
