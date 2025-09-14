package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gocql/gocql"
)

func getEMailListSelect(w http.ResponseWriter, r *http.Request) { // Выбирать все списки рассылки пользователя и активные email в них
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a list of emaillist", claims["username"])

	type ResponseEMailListDataT struct {
		ID          string `json:"id"`
		EMail       string `json:"email"`
		Description string `json:"description"`
		Status      int    `json:"status"`
	}

	type ResponseEMailListT struct {
		ID          string                   `json:"id"`
		Name        string                   `json:"name"`
		Description string                   `json:"description"`
		Status      int                      `json:"status"`
		Data        []ResponseEMailListDataT `json:"data"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseEmaiListData ResponseEMailListDataT
			var responseEMailList ResponseEMailListT
			var responseEMailListA []ResponseEMailListT

			ctx01 := context.Background()
			scanner01 := Session.Query(`SELECT id, name, description, status FROM emaillist WHERE user_id = ?`, claims["userid"]).WithContext(ctx01).Consistency(ConsistencyRead).Iter().Scanner()
			for scanner01.Next() {
				var emailListID string
				var emailListName string
				var emailListDescription string
				var emailListStatus int
				err := scanner01.Scan(&emailListID, &emailListName, &emailListDescription, &emailListStatus)
				if err != nil {
					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					var responseEMailListDataA []ResponseEMailListDataT
					ctx02 := context.Background()
					scanner02 := Session.Query(`SELECT id, email, description, status FROM emaillist_data WHERE emaillist_id = ?`, emailListID).WithContext(ctx02).Consistency(ConsistencyRead).Iter().Scanner()
					for scanner02.Next() {
						var emailListDataID string
						var email string
						var description string
						var status int
						err := scanner02.Scan(&emailListDataID, &email, &description, &status)
						if err != nil {
							log.Print(err)
							w.WriteHeader(http.StatusInternalServerError)
							return
						} else {
							responseEmaiListData.ID = emailListDataID
							responseEmaiListData.EMail = email
							responseEmaiListData.Description = description
							responseEmaiListData.Status = status

							responseEMailListDataA = append(responseEMailListDataA, responseEmaiListData)
						}
					}
					responseEMailList.ID = emailListID
					responseEMailList.Name = emailListName
					responseEMailList.Description = emailListDescription
					responseEMailList.Status = emailListStatus
					responseEMailList.Data = responseEMailListDataA

					responseEMailListA = append(responseEMailListA, responseEMailList)
				}
			}

			responseJSON, err := json.Marshal(responseEMailListA)
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

func putEMailListCreate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of a emaillis!", claims["username"])

	type EMailList struct {
		Name        string `json:"name"`
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

			id, err := gocql.RandomUUID()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var emailList EMailList
			if err := json.Unmarshal(b, &emailList); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`
			INSERT INTO emaillist (user_id, id, name, description, status) VALUES (?, ?, ?, ?, ?)
			`, claims["userid"], id, emailList.Name, emailList.Description, emailList.Status).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to create a emaillist (ERROR: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully created emaillist!", claims["username"])
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to create a emaillist!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putEMailListEdit(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested emaillist update!", claims["username"])

	type EMailListData struct {
		ID          string `json:"id"`
		EMail       string `json:"email"`
		Description string `json:"description"`
		Status      int    `json:"status"`
	}

	type EMailList struct {
		ID            string          `json:"id"`
		Name          string          `json:"name"`
		Description   string          `json:"description"`
		Status        int             `json:"status"`
		EMailListData []EMailListData `json:"data"`
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

			ctx := context.Background()

			var emailList EMailList
			if err := json.Unmarshal(b, &emailList); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = Session.Query(`DELETE FROM emaillist_data WHERE emaillist_id = ?`, emailList.ID).WithContext(ctx).Exec()
			if err != nil {
				log.Print("Failed to remove emaillist (ERROR: ", err.Error(), ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				err = Session.Query(`UPDATE emaillist SET name = ?, description = ?, status = ? WHERE user_id = ? AND id = ?`,
					emailList.Name, emailList.Description, emailList.Status, claims["userid"], emailList.ID).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to update emaillist (ERROR: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					for _, v := range emailList.EMailListData {
						uuidEMailListData, err := gocql.RandomUUID()
						if err != nil {
							w.WriteHeader(http.StatusInternalServerError)
							return
						}
						err = Session.Query(`
					INSERT INTO emaillist_data (emaillist_id, id, email, description, status) VALUES (?, ?, ?, ?, ?)
					`, emailList.ID, uuidEMailListData, v.EMail, v.Description, v.Status).WithContext(ctx).Exec()
						if err != nil {
							log.Print("Failed to update emaillist data (ERROR: ", err.Error(), ")!")
							w.WriteHeader(http.StatusInternalServerError)
							return
						}
					}
					log.Printf("User %v successfully created emaillist!", claims["username"])
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putEMailListRemove(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested emaillist removal!", claims["username"])

	type EMailList struct {
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

			var emailList EMailList
			if err := json.Unmarshal(b, &emailList); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err = Session.Query(`
			BEGIN BATCH
				DELETE FROM emaillist_data WHERE emaillist_id = ?
				DELETE FROM emaillist WHERE user_id = ? AND id = ?
			APPLY BATCH
			`, emailList.ID, claims["userid"], emailList.ID).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to remove emaillist (ERROR: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				log.Printf("User %v successfully deleted emaillist!", claims["username"])
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			log.Print("Failed to remove emaillist (API VERSION ERROR)!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}
