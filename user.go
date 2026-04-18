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

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested API userinfo!", claims["username"])

	type UserInfo struct {
		UserName string `json:"username"`
		FName    string `json:"fname"`
		LName    string `json:"lname"`
		UserID   string `json:"userid"`
		Role     string `json:"role"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		ctx := context.Background()
		scanner := Session.Query(`SELECT user_name, first_name, last_name, status, user_id, role FROM user WHERE user_name = ?`, claims["username"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
		for scanner.Next() {
			var userName string
			var firstName string
			var lastName string
			var status int
			var userID string
			var role string
			err := scanner.Scan(&userName, &firstName, &lastName, &status, &userID, &role)
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/userinfo", "404", "GET", "getUserInfo").Set(duration)

				log.Print("ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusNotFound)
				return
			} else {
				if status == 1 {
					userinfo := UserInfo{
						UserName: userName,
						FName:    firstName,
						LName:    lastName,
						UserID:   userID,
						Role:     role,
					}
					jsonAPI, err := json.Marshal(userinfo)
					if err != nil {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/userinfo", "500", "GET", "getUserInfo").Set(duration)

						log.Print("JSON ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/userinfo", "200", "GET", "getUserInfo").Set(duration)

					w.Header().Set("Content-Type", "application/json")
					w.Write(jsonAPI)
				} else {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/userinfo", "404", "GET", "getUserInfo").Set(duration)

					w.WriteHeader(http.StatusNotFound)
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/userinfo", "404", "GET", "getUserInfo").Set(duration)

			log.Print("ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

func putPassword(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a password change!", claims["username"])

	type Passwords struct {
		CurPassword string `json:"curpassword"`
		NewPassword string `json:"newpassword"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/password", "500", "PUT", "putPassword").Set(duration)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var passwords Passwords
		if err := json.Unmarshal(b, &passwords); err != nil {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/password", "500", "PUT", "putPassword").Set(duration)

			log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		{
			ctx := context.Background()
			scanner := Session.Query(`SELECT status, password FROM user WHERE user_name = ? LIMIT 1`, claims["username"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			for scanner.Next() {
				var status int
				var passwordDB string
				err := scanner.Scan(&status, &passwordDB)
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/password", "404", "PUT", "putPassword").Set(duration)

					log.Print("Failed to change the password (SELECT: ", err.Error(), ")!")
					w.WriteHeader(http.StatusNotFound)
					return
				} else {
					if passwordDB == passwords.CurPassword && status == 1 {
						err := Session.Query(`UPDATE user SET password = ? WHERE user_name = ?`, passwords.NewPassword, claims["username"]).WithContext(ctx).Exec()
						if err != nil {
							duration := time.Since(start).Seconds()
							httpDuration.WithLabelValues("/password", "500", "PUT", "putPassword").Set(duration)

							log.Print("Failed to change the password (UPDATE: ", err.Error(), ")!")
							w.WriteHeader(http.StatusInternalServerError)
							return
						} else {
							duration := time.Since(start).Seconds()
							httpDuration.WithLabelValues("/password", "200", "PUT", "putPassword").Set(duration)

							log.Printf("User %v successfully changed the password!", claims["username"])
							w.WriteHeader(http.StatusOK)
							return
						}
					}
				}
			}
		}
	}
	duration := time.Since(start).Seconds()
	httpDuration.WithLabelValues("/password", "404", "PUT", "putPassword").Set(duration)

	log.Print("It was not possible to change the user profile!")
	w.WriteHeader(http.StatusNotFound)
}

func putUserCreate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of the profile!", claims["username"])

	type User struct {
		UserName    string `json:"username"`
		Password    string `json:"password"`
		FirstName   string `json:"firstname"`
		LastName    string `json:"lastname"`
		Status      int    `json:"status"`
		Description string `json:"description"`
		Role        string `json:"role"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/create", "500", "PUT", "putUserCreate").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var user User
			if err := json.Unmarshal(b, &user); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/create", "500", "PUT", "putUserCreate").Set(duration)

				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				userID, err := gocql.RandomUUID()
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/user/create", "500", "PUT", "putUserCreate").Set(duration)

					log.Print("Failed to create an UUID user (", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if claims["role"] == "admin" {
					user.Role = "user"
				}
				ctx := context.Background()
				err = Session.Query(`
				BEGIN BATCH
					INSERT INTO user (user_id, user_name, password, first_name, last_name, status, description, role, owner_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
					INSERT INTO user_id (user_id, user_name) VALUES (?, ?)
					INSERT INTO owner_id_user (owner_id, user_name) VALUES (?, ?)
				APPLY BATCH`,
					userID, user.UserName, user.Password, user.FirstName, user.LastName, user.Status, user.Description, user.Role, claims["userid"], userID, user.UserName, claims["userid"], user.UserName).WithContext(ctx).Exec()
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/user/create", "500", "PUT", "putUserCreate").Set(duration)

					log.Print("Failed to create a user profile (INSERT: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/create", "200", "PUT", "putUserCreate").Set(duration)

				log.Printf("User %v successfully created a profile for the user %v!", claims["username"], user.UserName)
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/user/create", "500", "PUT", "putUserCreate").Set(duration)

			log.Print("It was not possible to create a profile!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/user/create", "403", "PUT", "putUserCreate").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putUserUpdate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a profile update!", claims["username"])

	type User struct {
		UserName    string `json:"username"`
		Password    string `json:"password"`
		FirstName   string `json:"firstname"`
		LastName    string `json:"lastname"`
		Status      int    `json:"status"`
		Description string `json:"description"`
		Role        string `json:"role"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/update", "500", "PUT", "putUserUpdate").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var user User
			if err := json.Unmarshal(b, &user); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/update", "500", "PUT", "putUserUpdate").Set(duration)

				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if claims["role"] == "admin" {
				user.Role = "user"
			}

			{
				ctx := context.Background()
				var err error
				if user.Password == "" {
					err = Session.Query(`UPDATE user SET first_name = ?, last_name = ?, status = ?, description = ?, role = ? WHERE user_name = ?`, user.FirstName, user.LastName, user.Status, user.Description, user.Role, user.UserName).WithContext(ctx).Exec()
				} else {
					err = Session.Query(`UPDATE user SET first_name = ?, last_name = ?, status = ?, description = ?, role = ?, password = ? WHERE user_name = ?`, user.FirstName, user.LastName, user.Status, user.Description, user.Role, user.Password, user.UserName).WithContext(ctx).Exec()
				}
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/user/update", "500", "PUT", "putUserUpdate").Set(duration)

					log.Print("Failed to change the user data (UPDATE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/user/update", "200", "PUT", "putUserUpdate").Set(duration)

					log.Printf("User %v successfully changed the user data %v!", claims["username"], user.UserName)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/user/update", "500", "PUT", "putUserUpdate").Set(duration)

			log.Print("It was not possible to change the data!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/user/update", "403", "PUT", "putUserUpdate").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putUserRemove(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a profile removal!", claims["username"])

	type User struct {
		UserName string `json:"username"`
		UserID   string `json:"userid"`
		OwnerID  string `json:"ownerid"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/remove", "500", "PUT", "putUserRemove").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var user User
			if err := json.Unmarshal(b, &user); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/remove", "500", "PUT", "putUserRemove").Set(duration)

				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`
				BEGIN BATCH
					DELETE FROM user WHERE user_name = ?
					DELETE FROM user_id WHERE user_id = ?
					DELETE FROM owner_id_user WHERE owner_id = ? AND user_name = ?
				APPLY BATCH`,
					user.UserName, user.UserID, user.OwnerID, user.UserName).WithContext(ctx).Exec()
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/user/remove", "500", "PUT", "putUserRemove").Set(duration)

					log.Print("Failed to delete user data (DELETE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/remove", "200", "PUT", "putUserRemove").Set(duration)

				log.Printf("User %v successfully deleted user profile %v!", claims["username"], user.UserName)
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/user/remove", "500", "PUT", "putUserRemove").Set(duration)

			log.Print("It was not possible to remove the profile!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/user/remove", "403", "PUT", "putUserRemove").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func getUserCount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		var userList []string

		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "admin" {
				scanner = Session.Query(`SELECT user_name FROM owner_id_user WHERE owner_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var userName string
					err := scanner.Scan(&userName)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						userList = append(userList, userName)
					}
				}
			}

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT count(*) FROM user`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else {
				args := []interface{}{}
				for _, item := range userList {
					args = append(args, item)
				}
				if len(userList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM user WHERE user_name IN (?`+strings.Repeat(", ?", len(userList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseCount.Count = 0
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

func putUserSelect(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	type ResponseUserDataT struct {
		UserID      string `json:"userid"`
		UserName    string `json:"username"`
		FirstName   string `json:"firstname"`
		LastName    string `json:"lastname"`
		Status      int    `json:"status"`
		LoginTime   int64  `json:"logintime"`
		Description string `json:"description"`
		Role        string `json:"role"`
		OwnerID     string `json:"ownerid"`
	}
	type ResponseUserT struct {
		Data        []ResponseUserDataT `json:"data"`
		CurrentPage string              `json:"current"`
		NextPage    string              `json:"next"`
	}

	type SelectParam struct {
		CurrentPage string `json:"current"`
		PageSize    int    `json:"pagesize"`
	}

	var userList []string

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			ctx := context.Background()
			var scanner gocql.Scanner
			var iter *gocql.Iter

			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/select", "500", "PUT", "putUserSelect").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var selectParam SelectParam
			if err := json.Unmarshal(b, &selectParam); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/select", "500", "PUT", "putUserSelect").Set(duration)

				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			pageSize := max(selectParam.PageSize, 0)

			var responseUserData ResponseUserDataT
			var responseUser ResponseUserT
			var pageState []byte

			if selectParam.CurrentPage != "" {
				var err error
				pageState, err = base64.StdEncoding.DecodeString(selectParam.CurrentPage)
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/user/select", "500", "PUT", "putUserSelect").Set(duration)

					log.Print("CurrentPage decode ERROR (" + err.Error() + ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			if claims["role"] == "admin" {
				scanner = Session.Query(`SELECT user_name FROM owner_id_user WHERE owner_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var userName string
					err := scanner.Scan(&userName)
					if err != nil {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/user/select", "500", "GET", "getUserSelect").Set(duration)

						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						userList = append(userList, userName)
					}
				}

				args := []interface{}{}
				for _, item := range userList {
					args = append(args, item)
				}
				if len(userList) > 0 {
					if selectParam.PageSize == 0 {
						scanner = Session.Query(`SELECT user_name, description, first_name, last_name, login_time, status, user_id, role, owner_id FROM user WHERE user_name IN (?`+strings.Repeat(", ?", len(userList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					} else {
						iter = Session.Query(`SELECT user_name, description, first_name, last_name, login_time, status, user_id, role, owner_id FROM user WHERE user_name IN (?`+strings.Repeat(", ?", len(userList)-1)+`)`, args...).PageSize(pageSize).PageState(pageState).Iter()
					}
				} else {
					responseUserJSON, err := json.Marshal(responseUser)
					if err != nil {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/user/select", "500", "GET", "getUserSelect").Set(duration)

						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/user/select", "200", "GET", "getUserSelect").Set(duration)

						w.Header().Set("Content-Type", "application/json")
						w.Write(responseUserJSON)
						return
					}
				}
			} else {
				if selectParam.PageSize == 0 {
					scanner = Session.Query(`SELECT user_name, description, first_name, last_name, login_time, status, user_id, role, owner_id FROM user`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					iter = Session.Query(`SELECT user_name, description, first_name, last_name, login_time, status, user_id, role, owner_id FROM user`).PageSize(pageSize).PageState(pageState).Iter()
				}
			}

			if selectParam.PageSize > 0 {
				nextPageState := iter.PageState()
				if len(nextPageState) > 0 {
					responseUser.NextPage = base64.StdEncoding.EncodeToString(nextPageState)
				} else {
					responseUser.NextPage = ""
				}
				responseUser.CurrentPage = selectParam.CurrentPage

				scanner = iter.Scanner()
			}

			for scanner.Next() {
				var userName string
				var description string
				var firstName string
				var lastName string
				var loginTime int64
				var status int
				var userID string
				var role string
				var ownerid string

				err := scanner.Scan(&userName, &description, &firstName, &lastName, &loginTime, &status, &userID, &role, &ownerid)
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/user/select", "500", "PUT", "putUserSelect").Set(duration)

					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseUserData.UserID = userID
					responseUserData.UserName = userName
					responseUserData.FirstName = firstName
					responseUserData.LastName = lastName
					responseUserData.LoginTime = loginTime
					responseUserData.Status = status
					responseUserData.Description = description
					responseUserData.Role = role
					responseUserData.OwnerID = ownerid

					responseUser.Data = append(responseUser.Data, responseUserData)
				}
			}
			if err := scanner.Err(); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/select", "500", "PUT", "putUserSelect").Set(duration)

				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseUserJSON, err := json.Marshal(responseUser)
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/select", "500", "PUT", "putUserSelect").Set(duration)

				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/user/select", "200", "PUT", "putUserSelect").Set(duration)

				w.Header().Set("Content-Type", "application/json")
				w.Write(responseUserJSON)
				return
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/user/select", "500", "PUT", "putUserSelect").Set(duration)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/user/select", "403", "PUT", "putUserSelect").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}
