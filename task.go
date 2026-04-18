package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	rand "math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gocql/gocql"
)

type ResponseTaskDataT struct {
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
	DMID        string `json:"dmid"`
}
type ResponseTaskT struct {
	Data        []ResponseTaskDataT `json:"data"`
	CurrentPage string              `json:"current"`
	NextPage    string              `json:"next"`
}

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

func putTaskSelect(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ctx := context.Background()
	var scannerTask gocql.Scanner
	var iter *gocql.Iter

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
				httpDuration.WithLabelValues("/task/select", "500", "PUT", "putTaskSelect").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var selectParam SelectParam
			if err := json.Unmarshal(b, &selectParam); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/task/select", "500", "PUT", "putTaskSelect").Set(duration)

				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			pageSize := max(selectParam.PageSize, 0)

			var responseTaskData ResponseTaskDataT
			var responseTask ResponseTaskT
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

			switch claims["role"] {
			case "superadmin":
				if selectParam.PageSize == 0 {
					scannerTask = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type, emaillist_id, dm_id FROM task`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					iter = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type, emaillist_id, dm_id FROM task`).PageSize(pageSize).PageState(pageState).Iter()
				}
			case "admin":
				var moduleList []string
				scannerUser := Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scannerUser.Next() {
					var registrationID string
					err := scannerUser.Scan(&registrationID)
					if err != nil {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/task/select", "500", "GET", "putTaskSelect").Set(duration)

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
					if selectParam.PageSize == 0 {
						scannerTask = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type, emaillist_id, dm_id FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					} else {
						iter = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type, emaillist_id, dm_id FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).PageSize(pageSize).PageState(pageState).Iter()
					}
				} else {
					responseTaskJSON, err := json.Marshal(responseTask)
					if err != nil {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/task/select", "500", "GET", "putTaskSelect").Set(duration)

						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/task/select", "200", "GET", "putTaskSelect").Set(duration)

						w.Header().Set("Content-Type", "application/json")
						w.Write(responseTaskJSON)
						return
					}
				}
			default:
				var moduleList []string
				scannerUser := Session.Query(`SELECT registration_id FROM user_registration WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scannerUser.Next() {
					var registrationID string
					err := scannerUser.Scan(&registrationID)
					if err != nil {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/task/select", "500", "GET", "putTaskSelect").Set(duration)

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
					if selectParam.PageSize == 0 {
						scannerTask = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type, emaillist_id, dm_id FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					} else {
						iter = Session.Query(`SELECT module_id, space_id, object, metric, status, critical, warning, interval, data_type, emaillist_id, dm_id FROM task WHERE module_id IN (?`+strings.Repeat(", ?", len(moduleList)-1)+`)`, args...).PageSize(pageSize).PageState(pageState).Iter()
					}

				} else {
					responseTaskJSON, err := json.Marshal(responseTask)
					if err != nil {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/task/select", "500", "GET", "putTaskSelect").Set(duration)

						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						duration := time.Since(start).Seconds()
						httpDuration.WithLabelValues("/task/select", "200", "GET", "putTaskSelect").Set(duration)

						w.Header().Set("Content-Type", "application/json")
						w.Write(responseTaskJSON)
						return
					}
				}
			}

			if selectParam.PageSize > 0 {
				nextPageState := iter.PageState()
				if len(nextPageState) > 0 {
					responseTask.NextPage = base64.StdEncoding.EncodeToString(nextPageState)
				} else {
					responseTask.NextPage = ""
				}
				responseTask.CurrentPage = selectParam.CurrentPage

				scannerTask = iter.Scanner()
			}

			for scannerTask.Next() {
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
				var dmID string

				err := scannerTask.Scan(&moduleID, &spaceID, &object, &metric, &status, &critical, &warning, &interval, &dataType, &emailListID, &dmID)
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/task/select", "500", "GET", "putTaskSelect").Set(duration)

					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseTaskData.ModuleID = moduleID
					responseTaskData.SpaceID = spaceID
					responseTaskData.Object = object
					responseTaskData.Metric = metric
					responseTaskData.Status = status
					responseTaskData.Critical = critical
					responseTaskData.Warning = warning
					responseTaskData.Interval = interval
					responseTaskData.DataType = dataType
					responseTaskData.EMailListID = emailListID
					responseTaskData.DMID = dmID

					scannerModule := Session.Query(`SELECT description FROM registration WHERE id = ?`, moduleID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					var moduleDesc string
					for scannerModule.Next() {
						err := scannerModule.Scan(&moduleDesc)
						if err != nil {
							log.Print(err)
						} else {
							responseTaskData.ModuleDesc = moduleDesc
						}
					}

					scannerSpace := Session.Query(`SELECT description FROM space WHERE id = ?`, spaceID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					var spaceDesc string
					for scannerSpace.Next() {
						err := scannerSpace.Scan(&spaceDesc)
						if err != nil {
							log.Print(err)
						} else {
							responseTaskData.SpaceDesc = spaceDesc
						}
					}

					responseTask.Data = append(responseTask.Data, responseTaskData)
				}
			}
			if err := scannerTask.Err(); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/task/select", "500", "GET", "putTaskSelect").Set(duration)

				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseTaskJSON, err := json.Marshal(responseTask)
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/task/select", "500", "GET", "putTaskSelect").Set(duration)

				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/task/select", "200", "GET", "putTaskSelect").Set(duration)

				w.Header().Set("Content-Type", "application/json")
				w.Write(responseTaskJSON)
				return
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/task/select", "500", "GET", "putTaskSelect").Set(duration)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/task/select", "403", "GET", "putTaskSelect").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putTaskUpdate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

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
		DMID        string `json:"dmid"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/task/update", "500", "PUT", "putTaskUpdate").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var task Task
			if err := json.Unmarshal(b, &task); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/task/update", "500", "PUT", "putTaskUpdate").Set(duration)

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
			var DMIDValue *string
			if task.DMID == "" {
				DMIDValue = nil
			} else {
				DMIDValue = &task.DMID
			}

			{
				ctx := context.Background()
				err := Session.Query(`UPDATE task SET status = ?, critical = ?, warning = ?, interval = ?, data_type = ?, emaillist_id = ?, dm_id = ? WHERE module_id = ? AND space_id = ? AND object = ? AND metric = ?`, task.Status, task.Critical, task.Warning, task.Interval, task.DataType, emailListIDValue, DMIDValue, task.ModuleID, task.SpaceID, task.Object, task.Metric).WithContext(ctx).Exec()
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/task/update", "500", "PUT", "putTaskUpdate").Set(duration)

					log.Print("Failed to change these tasks (UPDATE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/task/update", "200", "PUT", "putTaskUpdate").Set(duration)

					log.Printf("User %v successfully changed these tasks for the module %v!", claims["username"], task.ModuleID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/task/update", "500", "PUT", "putTaskUpdate").Set(duration)

			log.Print("It was not possible to change the data!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/task/update", "403", "PUT", "putTaskUpdate").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putTaskCreate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

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
		DMID        string `json:"dmid"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/task/create", "500", "PUT", "putTaskCreate").Set(duration)

				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var task Task
			if err := json.Unmarshal(b, &task); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/task/create", "500", "PUT", "putTaskCreate").Set(duration)

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
			var DMIDValue *string
			if task.DMID == "" {
				DMIDValue = nil
			} else {
				DMIDValue = &task.DMID
			}

			{
				ctx := context.Background()
				err := Session.Query(`INSERT INTO task (module_id, space_id, object, metric, status, int_id, critical, warning, interval, data_type, emaillist_id, ol_time, dm_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, toTimestamp(now()), ?)`, task.ModuleID, task.SpaceID, task.Object, task.Metric, task.Status, rand.Int(), task.Critical, task.Warning, task.Interval, task.DataType, emailListIDValue, DMIDValue).WithContext(ctx).Exec()
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/task/create", "500", "PUT", "putTaskCreate").Set(duration)

					log.Print("Failed to create a task (INSERT: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/task/create", "200", "PUT", "putTaskCreate").Set(duration)

					log.Printf("User %v successfully created the task!", claims["username"])
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/task/create", "500", "PUT", "putTaskCreate").Set(duration)

			log.Print("It was not possible to create a task!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/task/create", "403", "PUT", "putTaskCreate").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putTaskRemove(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

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
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/task/remove", "500", "PUT", "putTaskRemove").Set(duration)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var task Task
			if err := json.Unmarshal(b, &task); err != nil {
				duration := time.Since(start).Seconds()
				httpDuration.WithLabelValues("/task/remove", "500", "PUT", "putTaskRemove").Set(duration)

				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			{
				ctx := context.Background()
				err := Session.Query(`DELETE FROM task WHERE module_id = ? AND space_id = ? AND object = ? AND metric = ?`, task.ModuleID, task.SpaceID, task.Object, task.Metric).WithContext(ctx).Exec()
				if err != nil {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/task/remove", "500", "PUT", "putTaskRemove").Set(duration)

					log.Print("Failed to delete the task (DELETE: ", err.Error(), ")!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					duration := time.Since(start).Seconds()
					httpDuration.WithLabelValues("/task/remove", "200", "PUT", "putTaskRemove").Set(duration)

					log.Printf("User %v successfully deleted the module task %v!", claims["username"], task.ModuleID)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		} else {
			duration := time.Since(start).Seconds()
			httpDuration.WithLabelValues("/task/remove", "500", "PUT", "putTaskRemove").Set(duration)

			log.Print("It was not possible to delete the task!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues("/task/remove", "403", "PUT", "putTaskRemove").Set(duration)

		w.WriteHeader(http.StatusForbidden)
		return
	}
}
