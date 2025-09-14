package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gocql/gocql"
)

func getProblemCount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {

		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			switch claims["role"] {
			case "superadmin":
				scanner = Session.Query(`SELECT count(*) FROM problem`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			case "admin":
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						log.Print(err)
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
					scanner = Session.Query(`SELECT count(*) FROM problem WHERE module_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						log.Print(err)
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
					scanner = Session.Query(`SELECT count(*) FROM problem WHERE module_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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

func getProblemSelect(w http.ResponseWriter, r *http.Request) {
	type ResponseProblemT struct {
		ModuleID       string `json:"moduleid"`
		SpaceID        string `json:"spaceid"`
		Metric         string `json:"metric"`
		Status         string `json:"status"`
		EventTime      int64  `json:"eventtime"`
		EventTimeStart int64  `json:"eventtimestart"`
		Value          string `json:"value"`
		ModuleDesc     string `json:"moduledesc"`
		SpaceDesc      string `json:"spacedesc"`
	}
	type ResponseProblemAT []ResponseProblemT

	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {

			var responseProblem ResponseProblemT
			var responseProblemA ResponseProblemAT

			ctx := context.Background()
			var scanner gocql.Scanner

			switch claims["role"] {
			case "superadmin":
				scanner = Session.Query(`SELECT space_id, module_id, metric, event_time, event_time_start, status, value FROM problem`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			case "admin":
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						log.Print(err)
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
					scanner = Session.Query(`SELECT space_id, module_id, metric, event_time, event_time_start, status, value FROM problem WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				} else {
					responseProblemJSON, err := json.Marshal(responseProblemA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseProblemJSON)
						return
					}
				}
			default:
				var spaceList []string
				scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var spaceID string
					err := scanner.Scan(&spaceID)
					if err != nil {
						log.Print(err)
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
					scanner = Session.Query(`SELECT space_id, module_id, metric, event_time, event_time_start, status, value FROM problem WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Iter().Scanner()
				} else {
					responseProblemJSON, err := json.Marshal(responseProblemA)
					if err != nil {
						log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.Write(responseProblemJSON)
						return
					}
				}
			}

			for scanner.Next() {
				var moduleID string
				var spaceID string
				var metric string
				var status string
				var eventTime int64
				var eventTimeStart int64
				var value string

				err := scanner.Scan(&spaceID, &moduleID, &metric, &eventTime, &eventTimeStart, &status, &value)
				if err != nil {
					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					responseProblem.ModuleID = moduleID
					responseProblem.SpaceID = spaceID
					responseProblem.Metric = metric
					responseProblem.Status = status
					responseProblem.EventTime = eventTime
					responseProblem.EventTimeStart = eventTimeStart
					responseProblem.Value = value

					scannerModule := Session.Query(`SELECT description FROM registration WHERE id = ?`, moduleID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					var moduleDesc string
					for scannerModule.Next() {
						err := scannerModule.Scan(&moduleDesc)
						if err != nil {
							log.Print(err)
						} else {
							responseProblem.ModuleDesc = moduleDesc
						}
					}

					scannerSpace := Session.Query(`SELECT description FROM space WHERE id = ?`, spaceID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					var spaceDesc string
					for scannerSpace.Next() {
						err := scannerSpace.Scan(&spaceDesc)
						if err != nil {
							log.Print(err)
						} else {
							responseProblem.SpaceDesc = spaceDesc
						}
					}

					responseProblemA = append(responseProblemA, responseProblem)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseProblemJSON, err := json.Marshal(responseProblemA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseProblemJSON)
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
