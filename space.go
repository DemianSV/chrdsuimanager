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

func getSpaceCount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			switch claims["role"] {
			case "superadmin":
				scanner = Session.Query(`SELECT count(*) FROM space`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			case "admin":
				scanner = Session.Query(`SELECT count(*) FROM user_space WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			default:
				scanner = Session.Query(`SELECT count(*) FROM user_space WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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

func getSpaceSelect(w http.ResponseWriter, r *http.Request) {
	type ResponseT struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		Status      int    `json:"status"`
		UserID      string `json:"userid"`
	}
	type ResponseAT []ResponseT

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {

			var response ResponseT
			var responseA ResponseAT

			ctx := context.Background()
			var scanner gocql.Scanner

			switch claims["role"] {
			case "superadmin":
				scanner = Session.Query(`SELECT id, description, status, user_id FROM space`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			case "admin":
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						spaceList = append(spaceList, spaceID)
					}
				}

				args := []interface{}{}
				for _, item := range spaceList {
					args = append(args, item)
				}
				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT id, description, status, user_id FROM space WHERE id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseJSON, err := json.Marshal(responseA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseJSON)
						return
					}
				}
			default:
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						spaceList = append(spaceList, spaceID)
					}
				}

				args := []interface{}{}
				for _, item := range spaceList {
					args = append(args, item)
				}
				if len(spaceList) > 0 {
					scanner = Session.Query(`SELECT id, description, status, user_id FROM space WHERE id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseJSON, err := json.Marshal(responseA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseJSON)
						return
					}
				}
			}
			for scanner.Next() {
				var id string
				var description string
				var status int
				var userID string

				err := scanner.Scan(&id, &description, &status, &userID)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					response.ID = id
					response.Description = description
					response.Status = status
					response.UserID = userID

					responseA = append(responseA, response)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseJSON, err := json.Marshal(responseA)
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

func putSpaceCreate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of space!", claims["username"])

	type Space struct {
		Description string `json:"description"`
		Status      int    `json:"status"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			id, err := gocql.RandomUUID()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var space Space
			if err := json.Unmarshal(b, &space); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`
				BEGIN BATCH
					INSERT INTO space (id, description, status, user_id) VALUES (?, ?, ?, ?)
					INSERT INTO user_space (user_id, space_id) VALUES (?, ?)
				APPLY BATCH
				`, id, space.Description, space.Status, claims["userid"], claims["userid"], id).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to create a space (INSERT: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				log.Printf("User %v successfully created the space!", claims["username"])
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			log.Print("It was not possible to create a space!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putSpaceUpdate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the update of space!", claims["username"])

	type Space struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		Status      int    `json:"status"`
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

			var space Space
			log.Print(string(b))
			if err := json.Unmarshal(b, &space); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`UPDATE space SET description = ?, status = ? WHERE id = ?`, space.Description, space.Status, space.ID).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to change these space (UPDATE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully changed these spaces %v!", claims["username"], space.ID)
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

func putSpaceRemove(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the removal of space!", claims["username"])

	type Space struct {
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

			var space Space
			if err := json.Unmarshal(b, &space); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				var userID string
				if err := Session.Query(`SELECT user_id FROM space WHERE id = ? LIMIT 1`, space.ID).WithContext(ctx).Consistency(ConsistencyRead).Scan(&userID); err != nil {
					log.Print("Failed to remove space (SELECT user_id: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
				}

				err := Session.Query(`
				BEGIN BATCH 
					DELETE FROM space WHERE id = ?
					DELETE FROM user_space WHERE user_id = ? AND space_id = ?
				APPLY BATCH
			`, space.ID, userID, space.ID).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to remove space (DELETE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				log.Printf("User %v successfully removed the space %v!", claims["username"], space.ID)
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			log.Print("It was not possible to remove the space!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}
