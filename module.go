package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gocql/gocql"
)

type ResponseModuleDataT struct {
	ID          string `json:"id"`
	TypeMod     string `json:"type"`
	Status      int    `json:"status"`
	Description string `json:"description"`
	UserID      string `json:"userid"`
	Address     string `json:"address"`
}
type ResponseModuleAT []ResponseModuleDataT
type ResponseModuleT struct {
	Data        []ResponseModuleDataT `json:"data"`
	CurrentPage string                `json:"current"`
	NextPage    string                `json:"next"`
}

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

func putModuleSelect(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	type SelectParam struct {
		CurrentPage string `json:"current"`
		PageSize    int    `json:"pagesize"`
	}

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/select", "500", "PUT", "putModuleSelect").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var selectParam SelectParam
			if err := json.Unmarshal(b, &selectParam); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/select", "500", "PUT", "putTaskSelect").Set(duration)

				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			pageSize := max(selectParam.PageSize, 0)

			var responseModuleData ResponseModuleDataT
			var responseModule ResponseModuleT
			var pageState []byte

			if selectParam.CurrentPage != "" {
				var err error
				pageState, err = base64.StdEncoding.DecodeString(selectParam.CurrentPage)
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/task/select", "500", "PUT", "putTaskSelect").Set(duration)

					log.Print("CurrentPage decode ERROR (" + err.Error() + ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			ctx := context.Background()
			var scannerReg gocql.Scanner
			var iter *gocql.Iter

			switch claims["role"] {
			case "superadmin":
				if selectParam.PageSize == 0 {
					scannerReg = Session.Query(`SELECT id, type, status, description, user_id, address FROM registration`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					iter = Session.Query(`SELECT id, type, status, description, user_id, address FROM registration`).PageSize(pageSize).PageState(pageState).Iter()
				}
			case "admin":
				var registrationList []string
				scannerUser := Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scannerUser.Next() {
					var registrationID string
					err := scannerUser.Scan(&registrationID)
					if err != nil {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/registration/select", "500", "GET", "putModuleSelect").Set(duration)

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
					if selectParam.PageSize == 0 {
						scannerReg = Session.Query(`SELECT id, type, status, description, user_id, address FROM registration WHERE id IN (?`+strings.Repeat(", ?", len(registrationList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					} else {
						iter = Session.Query(`SELECT id, type, status, description, user_id, address FROM registration WHERE id IN (?`+strings.Repeat(", ?", len(registrationList)-1)+`)`, args...).PageSize(pageSize).PageState(pageState).Iter()
					}
				} else {
					responseModuleJSON, err := json.Marshal(responseModule)
					if err != nil {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/registration/select", "500", "GET", "putModuleSelect").Set(duration)

						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/registration/select", "200", "GET", "putModuleSelect").Set(duration)

						w.Header().Set("Content-Type", "application/json")
						w.Write(responseModuleJSON)
						return
					}
				}
			default:
				var registrationList []string
				scannerUser := Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Iter().Scanner()
				for scannerUser.Next() {
					var registrationID string
					err := scannerUser.Scan(&registrationID)
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
					if selectParam.PageSize == 0 {
						scannerReg = Session.Query(`SELECT id, type, status, description, user_id, address FROM registration WHERE id IN (?`+strings.Repeat(", ?", len(registrationList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					} else {
						iter = Session.Query(`SELECT id, type, status, description, user_id, address FROM registration WHERE id IN (?`+strings.Repeat(", ?", len(registrationList)-1)+`)`, args...).PageSize(pageSize).PageState(pageState).Iter()
					}
				} else {
					responseModuleJSON, err := json.Marshal(responseModule)
					if err != nil {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/registration/select", "500", "GET", "putModuleSelect").Set(duration)

						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/registration/select", "200", "GET", "putModuleSelect").Set(duration)

						w.Header().Set("Content-Type", "application/json")
						w.Write(responseModuleJSON)
						return
					}
				}
			}

			if selectParam.PageSize > 0 {
				nextPageState := iter.PageState()
				if len(nextPageState) > 0 {
					responseModule.NextPage = base64.StdEncoding.EncodeToString(nextPageState)
				} else {
					responseModule.NextPage = ""
				}
				responseModule.CurrentPage = selectParam.CurrentPage

				scannerReg = iter.Scanner()
			}

			for scannerReg.Next() {
				var id string
				var description string
				var typeMod string
				var status int
				var userID string
				var address string

				err := scannerReg.Scan(&id, &typeMod, &status, &description, &userID, &address)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseModuleData.ID = id
					responseModuleData.TypeMod = typeMod
					responseModuleData.Status = status
					responseModuleData.Description = description
					responseModuleData.UserID = userID
					responseModuleData.Address = address

					responseModule.Data = append(responseModule.Data, responseModuleData)
				}
			}
			if err := scannerReg.Err(); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/select", "500", "GET", "putModuleSelect").Set(duration)

				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseModuleJSON, err := json.Marshal(responseModule)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/select", "200", "GET", "putModuleSelect").Set(duration)

				w.Header().Set("Content-Type", "application/json")
				w.Write(responseModuleJSON)
				return
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/registration/select", "500", "GET", "putModuleSelect").Set(duration)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/registration/select", "403", "GET", "putModuleSelect").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putModuleUpdate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the module update!", claims["username"])

	type Registration struct {
		ID          string `json:"id"`
		TypeMod     string `json:"type"`
		Address     string `json:"address"`
		Status      int    `json:"status"`
		Description string `json:"description"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/update", "500", "PUT", "putModuleUpdate").Set(duration)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var registration Registration
			log.Print(string(b))
			if err := json.Unmarshal(b, &registration); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/update", "500", "PUT", "putModuleUpdate").Set(duration)

				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`UPDATE registration SET type = ?, status = ?, description = ?, address = ? WHERE id = ?`, registration.TypeMod, registration.Status, registration.Description, registration.Address, registration.ID).WithContext(ctx).Exec()
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/registration/update", "500", "PUT", "putModuleUpdate").Set(duration)

					log.Print("Failed to change the module data (UPDATE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/registration/update", "200", "PUT", "putModuleUpdate").Set(duration)

					log.Printf("User %v successfully changed the module data %v!", claims["username"], registration.ID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/registration/update", "500", "PUT", "putModuleUpdate").Set(duration)

			log.Print("It was not possible to change the data!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/registration/update", "403", "PUT", "putModuleUpdate").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putModuleCreate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of the module!", claims["username"])

	type Registration struct {
		ID          string `json:"id"`
		TypeMod     string `json:"type"`
		Address     string `json:"address"`
		Status      int    `json:"status"`
		Description string `json:"description"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/create", "500", "PUT", "putModuleCreate").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			id, err := gocql.RandomUUID()
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/create", "500", "PUT", "putModuleCreate").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var registration Registration
			if err := json.Unmarshal(b, &registration); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/create", "500", "PUT", "putModuleCreate").Set(duration)

				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`
				BEGIN BATCH
					INSERT INTO registration (id, type, status, description, user_id, address) VALUES (?, ?, ?, ?, ?, ?)
					INSERT INTO user_registration (user_id, registration_id) VALUES (?, ?)
				APPLY BATCH`,
					id, registration.TypeMod, registration.Status, registration.Description, claims["userid"], registration.Address, claims["userid"], id).WithContext(ctx).Exec()
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/registration/create", "500", "PUT", "putModuleCreate").Set(duration)

					log.Print("Failed to create a module (BATCH: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/registration/create", "200", "PUT", "putModuleCreate").Set(duration)

					log.Printf("User %v successfully created a module!", claims["username"])
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/registration/create", "500", "PUT", "putModuleCreate").Set(duration)

			log.Print("It was not possible to create a module!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/registration/create", "403", "PUT", "putModuleCreate").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putModuleRemove(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

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
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/remove", "500", "PUT", "putModuleRemove").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var registration Registration
			if err := json.Unmarshal(b, &registration); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/registration/remove", "500", "PUT", "putModuleRemove").Set(duration)

				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()

				var userID string
				if err := Session.Query(`SELECT user_id FROM registration WHERE id = ? LIMIT 1`, registration.ID).WithContext(ctx).Consistency(ConsistencyRead).Scan(&userID); err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/registration/remove", "500", "PUT", "putModuleRemove").Set(duration)

					log.Print("Failed to remove the module (SELECT user_id: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				err := Session.Query(`
				BEGIN BATCH
					DELETE FROM registration WHERE id = ?
					DELETE FROM user_registration WHERE user_id = ? AND registration_id = ?
					DELETE FROM task WHERE module_id = ?
				APPLY BATCH`,
					registration.ID, userID, registration.ID, registration.ID).WithContext(ctx).Exec()
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/registration/remove", "500", "PUT", "putModuleRemove").Set(duration)

					log.Print("Failed to remove the module (DELETE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/registration/remove", "200", "PUT", "putModuleRemove").Set(duration)

					log.Printf("User %v successfully deleted the module %v!", claims["username"], registration.ID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/registration/remove", "500", "PUT", "putModuleRemove").Set(duration)

			log.Print("It was not possible to remove the module!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/registration/remove", "403", "PUT", "putModuleRemove").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}
