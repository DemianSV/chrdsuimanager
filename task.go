package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	rand "math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gocql/gocql"
)

type ResponseTaskT struct {
	ID          string `json:"id"`
	ModuleID    string `json:"moduleid"`
	SpaceID     string `json:"spaceid"`
	Object      string `json:"object"`
	Metric      string `json:"metric"`
	Status      int    `json:"status"`
	Critical    string `json:"critical"`
	Warning     string `json:"warning"`
	Interval    int64  `json:"interval"`
	DataType    string `json:"datatype"`
	ModuleDesc  string `json:"moduledesc"`
	SpaceDesc   string `json:"spacedesc"`
	EMailListID string `json:"emaillistid"`
}
type ResponseTaskAT []ResponseTaskT

func getTaskCount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			switch claims["role"] {
			case "superadmin":
				scanner = Session.Query(`SELECT count(*) FROM task`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			case "admin":
				var moduleList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						moduleList = append(moduleList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range moduleList {
					args = append(args, item)
				}

				if len(moduleList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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
			default:
				var moduleList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						moduleList = append(moduleList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range moduleList {
					args = append(args, item)
				}

				if len(moduleList) > 0 {
					scanner = Session.Query(`SELECT count(*) FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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
					log.Print(err)
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

func getTaskSelect(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {

			var responseTask ResponseTaskT
			var responseTaskA ResponseTaskAT

			ctx := context.Background()
			var scanner gocql.Scanner

			switch claims["role"] {
			case "superadmin":
				scanner = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type, emaillist_id FROM task`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			case "admin":
				var moduleList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						moduleList = append(moduleList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range moduleList {
					args = append(args, item)
				}

				if len(moduleList) > 0 {
					scanner = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type, emaillist_id FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseTaskJSON, err := json.Marshal(responseTaskA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseTaskJSON)
						return
					}
				}
			default:
				var moduleList []string
				scanner = Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var registrationID string
					err := scanner.Scan(&registrationID)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						moduleList = append(moduleList, registrationID)
					}
				}

				args := []interface{}{}
				for _, item := range moduleList {
					args = append(args, item)
				}

				if len(moduleList) > 0 {
					scanner = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type, emaillist_id FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseTaskJSON, err := json.Marshal(responseTaskA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseTaskJSON)
						return
					}
				}
			}

			for scanner.Next() {
				var moduleID string
				var spaceID string
				var object string
				var metric string
				var status int
				var critical string
				var warning string
				var interval int64
				var dataType string
				var emailListID string

				err := scanner.Scan(&moduleID, &spaceID, &object, &metric, &status, &critical, &warning, &interval, &dataType, &emailListID)
				if err != nil {
					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseTask.ModuleID = moduleID
					responseTask.SpaceID = spaceID
					responseTask.Object = object
					responseTask.Metric = metric
					responseTask.Status = status
					responseTask.Critical = critical
					responseTask.Warning = warning
					responseTask.Interval = interval
					responseTask.DataType = dataType
					responseTask.EMailListID = emailListID

					scannerModule := Session.Query(`SELECT description FROM registration WHERE id = ?`, moduleID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					var moduleDesc string
					for scannerModule.Next() {
						err := scannerModule.Scan(&moduleDesc)
						if err != nil {
							log.Print(err)
						} else {
							responseTask.ModuleDesc = moduleDesc
						}
					}

					scannerSpace := Session.Query(`SELECT description FROM space WHERE id = ?`, spaceID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					var spaceDesc string
					for scannerSpace.Next() {
						err := scannerSpace.Scan(&spaceDesc)
						if err != nil {
							log.Print(err)
						} else {
							responseTask.SpaceDesc = spaceDesc
						}
					}

					responseTaskA = append(responseTaskA, responseTask)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseTaskJSON, err := json.Marshal(responseTaskA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseTaskJSON)
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

func putTaskUpdate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the update of the problem!", claims["username"])

	type Task struct {
		ModuleID    string `json:"moduleid"`
		SpaceID     string `json:"spaceid"`
		Object      string `json:"object"`
		Metric      string `json:"metric"`
		Status      int    `json:"status"`
		Critical    string `json:"critical"`
		Warning     string `json:"warning"`
		Interval    int64  `json:"interval"`
		DataType    string `json:"datatype"`
		EMailListID string `json:"emaillistid"`
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

			var task Task
			if err := json.Unmarshal(b, &task); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var emailListIDValue *string
			if task.EMailListID == "" {
				emailListIDValue = nil
			} else {
				emailListIDValue = &task.EMailListID
			}

			{
				ctx := context.Background()
				err := Session.Query(`UPDATE task SET status = ?, critical = ?, warning = ?, interval = ?, data_type = ?, emaillist_id = ? WHERE module_id = ? AND space_id = ? AND object = ? AND metric = ?`, task.Status, task.Critical, task.Warning, task.Interval, task.DataType, emailListIDValue, task.ModuleID, task.SpaceID, task.Object, task.Metric).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to change these tasks (UPDATE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully changed these tasks for the module %v!", claims["username"], task.ModuleID)
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

func putTaskCreate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of the task!", claims["username"])

	type Task struct {
		ModuleID    string `json:"moduleid"`
		SpaceID     string `json:"spaceid"`
		Object      string `json:"object"`
		Metric      string `json:"metric"`
		Status      int    `json:"status"`
		Critical    string `json:"critical"`
		Warning     string `json:"warning"`
		Interval    int64  `json:"interval"`
		DataType    string `json:"datatype"`
		EMailListID string `json:"emaillistid"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var task Task
			if err := json.Unmarshal(b, &task); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var emailListIDValue *string
			if task.EMailListID == "" {
				emailListIDValue = nil
			} else {
				emailListIDValue = &task.EMailListID
			}

			{
				ctx := context.Background()
				err := Session.Query(`INSERT INTO task (module_id, space_id, object, metric, status, int_id, critical, warning, interval, data_type, emaillist_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, task.ModuleID, task.SpaceID, task.Object, task.Metric, task.Status, rand.Int(), task.Critical, task.Warning, task.Interval, task.DataType, emailListIDValue).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to create a task (INSERT: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully created the task!", claims["username"])
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to create a task!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putTaskRemove(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the removal of the task!", claims["username"])

	type Task struct {
		ModuleID string `json:"moduleid"`
		SpaceID  string `json:"spaceid"`
		Object   string `json:"object"`
		Metric   string `json:"metric"`
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

			var task Task
			if err := json.Unmarshal(b, &task); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`DELETE FROM task WHERE module_id = ? AND space_id = ? AND object = ? AND metric = ?`, task.ModuleID, task.SpaceID, task.Object, task.Metric).WithContext(ctx).Exec()
				if err != nil {
					log.Print("Failed to delete the task (DELETE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					log.Printf("User %v successfully deleted the module task %v!", claims["username"], task.ModuleID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			log.Print("It was not possible to delete the task!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}
