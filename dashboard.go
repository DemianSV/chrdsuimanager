package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/DemianSV/chrdsclient"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gocql/gocql"
)

func getDashboardCount(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		if versionAPI == "1" {
			var responseCount ResponseCountT
			ctx := context.Background()
			var scanner gocql.Scanner

			if claims["role"] == "superadmin" {
				scanner = Session.Query(`SELECT count(*) FROM dashboard WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			} else {
				scanner = Session.Query(`SELECT count(*) FROM dashboard WHERE user_id = ?`, claims["userid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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

func getDashboardSelect(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested a list of dashboard", claims["username"])

	type ResponseChartDataT struct {
		ID         string `json:"id"`
		Metric     string `json:"metric"`
		SpaceID    string `json:"spaceid"`
		GraphType  string `json:"graphtype"`
		GraphColor string `json:"graphcolor"`
		GroupFunc  string `json:"groupfunc"` // Max, Min, Avg, Sum
		GraphOrder int    `json:"graphorder"`
	}

	type ResponseDashboardT struct {
		ID        string               `json:"id"`
		Name      string               `json:"name"`
		StartTime int64                `json:"starttime"`
		StopTime  int64                `json:"stoptime"`
		Status    int                  `json:"status"`
		ChartData []ResponseChartDataT `json:"chartdata"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		var responseChartData ResponseChartDataT
		var responseDashboard ResponseDashboardT
		var responseDashboardA []ResponseDashboardT

		ctx01 := context.Background()
		scanner01 := Session.Query(`SELECT id, name, start_time, stop_time, status FROM dashboard WHERE user_id = ?`, claims["userid"]).WithContext(ctx01).Consistency(ConsistencyRead).Iter().Scanner()
		for scanner01.Next() {
			var dashboardID string
			var dashboardName string
			var startTime int64
			var stopTime int64
			var dashboardStatus int
			err := scanner01.Scan(&dashboardID, &dashboardName, &startTime, &stopTime, &dashboardStatus)
			if err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				if dashboardStatus == 1 {
					var responseChartDataA []ResponseChartDataT
					ctx02 := context.Background()
					scanner02 := Session.Query(`SELECT id, graph_color, graph_func, graph_type, metric, space_id, status, graph_order FROM dashboard_chartdata WHERE dashboard_id = ?`, dashboardID).WithContext(ctx02).Consistency(ConsistencyRead).Iter().Scanner()
					for scanner02.Next() {
						var chartDataID string
						var graphColor string
						var graphFunc string
						var graphType string
						var metric string
						var spaceID string
						var chartDataStatus int
						var graphOrder int
						err := scanner02.Scan(&chartDataID, &graphColor, &graphFunc, &graphType, &metric, &spaceID, &chartDataStatus, &graphOrder)
						if err != nil {
							log.Print(err)
							w.WriteHeader(http.StatusInternalServerError)
							return
						} else {
							responseChartData.ID = chartDataID
							responseChartData.Metric = metric
							responseChartData.SpaceID = spaceID
							responseChartData.GraphColor = graphColor
							responseChartData.GraphType = graphType
							responseChartData.GroupFunc = graphFunc
							responseChartData.GraphOrder = graphOrder

							responseChartDataA = append(responseChartDataA, responseChartData)
						}
					}
					responseDashboard.ID = dashboardID
					responseDashboard.Name = dashboardName
					responseDashboard.StartTime = startTime
					responseDashboard.StopTime = stopTime
					responseDashboard.Status = dashboardStatus
					responseDashboard.ChartData = responseChartDataA

					responseDashboardA = append(responseDashboardA, responseDashboard)
				}
			}
		}

		responseJSON, err := json.Marshal(responseDashboardA)
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
}

func putChartDashboardData(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested Dashboard data", claims["username"])

	type RequesChartDataT struct {
		ID         string `json:"id"`
		Metric     string `json:"metric"`
		SpaceID    string `json:"spaceid"`
		GraphType  string `json:"graphtype"`
		GraphColor string `json:"graphcolor"`
		GroupFunc  string `json:"groupfunc"` // Max, Min, Avg, Sum
		GraphOrder int    `json:"graphorder"`
	}

	type RequestChartDataSetT struct {
		ID        string             `json:"id"`
		Name      string             `json:"name"`
		StartTime int64              `json:"starttime"`
		StopTime  int64              `json:"stoptime"`
		ChartData []RequesChartDataT `json:"chartdata"`
	}

	type ChartDataSetT struct {
		Label           string    `json:"label"`
		Type            string    `json:"type"`
		BackgroundColor string    `json:"backgroundColor"`
		BorderColor     string    `json:"borderColor"`
		Data            []float32 `json:"data"`
		BarPercentage   float32   `json:"barPercentage"` // The space between the bars
		Order           int       `json:"order"`
	}

	type ChartDataT struct {
		Labels    []int64         `json:"labels"`
		DataSets  []ChartDataSetT `json:"datasets"`
		StartTime int64           `json:"starttime"`
		StopTime  int64           `json:"stoptime"`
	}

	if claims["role"] == "superadmin" || claims["role"] == "admin" || claims["role"] == "user" {
		versionAPI := chi.URLParam(r, "version")
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var request []RequestChartDataSetT
			if err := json.Unmarshal(b, &request); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var chartDataA []ChartDataT

			for _, requestA := range request {
				var chartDataSet ChartDataSetT
				var chartData ChartDataT

				var stopTime int64 = 0   // The end of the sample period
				var startTime int64 = 0  // The start time of the sample period
				var deltaTime int64 = 60 // Group step by sampling time in seconds
				ctx := context.Background()
				if requestA.StopTime == 0 {
					stopTime = chrdsclient.MakeTimestamp() - (60 * 1000)
				} else {
					stopTime = requestA.StopTime
				}
				if requestA.StartTime == 0 {
					startTime = stopTime - (30 * 60 * 1000) // Default sample period 30 minutes
				} else {
					startTime = requestA.StartTime
				}

				log.Print(startTime, " : ", stopTime)

				if stopTime-startTime <= (3 * 60 * 1000) { // Up to 3 minutes
					deltaTime = 1 // 1 second
				} else if stopTime-startTime > (3*60*1000) && stopTime-startTime <= (60*60*1000) { // up to 1 hour
					deltaTime = 60 // 1 minute
				} else if stopTime-startTime > (60*60*1000) && stopTime-startTime <= (120*60*1000) { // Up to 2 hours
					deltaTime = 600 // 10 minutes
				} else if stopTime-startTime > (120*60*1000) && stopTime-startTime <= (1440*60*1000) { // Up to 24 hours
					deltaTime = 900 // 15 minutes
				} else if stopTime-startTime > (1440*60*1000) && stopTime-startTime <= (4320*60*1000) { // Up to 3 days
					deltaTime = 3600 // 1 time
				} else { // More than 3 days
					deltaTime = 86400 // 1 day
				}

				chartData.StartTime = startTime
				chartData.StopTime = stopTime

				for iTime := startTime; iTime <= stopTime; iTime = iTime + (deltaTime * 1000) {
					chartData.Labels = append(chartData.Labels, iTime)
				}

				for _, reqChartData := range requestA.ChartData {
					chartDataSet.Label = reqChartData.Metric
					chartDataSet.Type = reqChartData.GraphType
					if reqChartData.GraphColor == "" {
						chartDataSet.BackgroundColor = "#ade2ffbe" // Default color: #ade2ffbe  ade2ffbe
					} else {
						chartDataSet.BackgroundColor = reqChartData.GraphColor
					}
					chartDataSet.BorderColor = reqChartData.GraphColor
					chartDataSet.BarPercentage = 0.9 // The space between the bars
					chartDataSet.Order = reqChartData.GraphOrder
					chartDataSet.Data = nil

					for _, iTime := range chartData.Labels {
						args := []interface{}{}
						args = append(args, reqChartData.SpaceID)
						args = append(args, reqChartData.Metric)
						args = append(args, iTime)
						args = append(args, iTime+(deltaTime*1000))

						syntKeyList := makeDateList(int(((stopTime - startTime) / 1000 / 60 / 60 / 24 / 30) + 1)) // Preparation of the list of keys for data sample
						for _, item := range syntKeyList {
							args = append(args, item)
						}

						scanner := Session.Query(`SELECT value FROM raw_data02 WHERE space_id = ? AND metric = ? AND event_time >= ? AND event_time < ? AND synt_key IN (?`+strings.Repeat(", ?", len(syntKeyList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()

						var valueCount float32 = 0
						var valueSum float32 = 0
						for scanner.Next() {
							var value float32
							err := scanner.Scan(&value)
							if err != nil {
								go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
								w.WriteHeader(http.StatusInternalServerError)
								return
							} else {
								valueSum = valueSum + value
							}
							valueCount++
						}
						if err := scanner.Err(); err != nil {
							log.Print(err)
							go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
							w.WriteHeader(http.StatusInternalServerError)
							return
						}

						if valueCount == 0 {
							chartDataSet.Data = append(chartDataSet.Data, 0)
						} else {
							chartDataSet.Data = append(chartDataSet.Data, valueSum/valueCount)
						}
					}

					chartData.DataSets = append(chartData.DataSets, chartDataSet)
				}

				chartDataA = append(chartDataA, chartData)
			}

			responseJSON, err := json.Marshal(chartDataA)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				go chrdsclient.Metric("httpstatus", float32(http.StatusOK))
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseJSON)
				return
			}
			go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Print("Failed to get data!")
			go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func putDashboardCreate(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested the creation of a dashboard!", claims["username"])

	type Dashboard struct {
		Name      string `json:"name"`
		StartTime int64  `json:"starttime"`
		StopTime  int64  `json:"stoptime"`
		Status    int    `json:"status"`
	}

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

		var dashboard Dashboard
		if err := json.Unmarshal(b, &dashboard); err != nil {
			log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		{
			ctx := context.Background()
			err := Session.Query(`
			INSERT INTO dashboard (user_id, id, name, start_time, stop_time, status) VALUES (?, ?, ?, ?, ?, ?)
			`, claims["userid"], id, dashboard.Name, dashboard.StartTime, dashboard.StopTime, dashboard.Status).WithContext(ctx).Exec()
			if err != nil {
				log.Print("Failed to create a dashboard (ERROR: ", err.Error(), ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				log.Printf("User %v successfully created Dashboard!", claims["username"])
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	} else {
		log.Print("It was not possible to create a dashboard!")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func putDashboardEdit(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested dashboard update!", claims["username"])

	type DashboardChartData struct {
		ID         string `json:"id"`
		Metric     string `json:"metric"`
		SpaceID    string `json:"spaceid"`
		GraphType  string `json:"graphtype"`
		GraphColor string `json:"graphcolor"`
		GroupFunc  string `json:"groupfunc"` // Max, Min, Avg, Sum
		GraphOrder int    `json:"graphorder"`
	}

	type Dashboard struct {
		ID        string               `json:"id"`
		Name      string               `json:"name"`
		StartTime int64                `json:"starttime"`
		StopTime  int64                `json:"stoptime"`
		Status    int                  `json:"status"`
		ChartData []DashboardChartData `json:"chartdata"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := context.Background()

		var dashboard Dashboard
		if err := json.Unmarshal(b, &dashboard); err != nil {
			log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = Session.Query(`
		BEGIN BATCH
			DELETE FROM dashboard_chartdata WHERE dashboard_id = ?
			DELETE FROM dashboard WHERE user_id = ? AND id = ?
		APPLY BATCH
		`, dashboard.ID, claims["userid"], dashboard.ID).WithContext(ctx).Exec()
		if err != nil {
			log.Print("Failed to remove dashboard (ERROR: ", err.Error(), ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			uuid, err := gocql.RandomUUID()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = Session.Query(`
			INSERT INTO dashboard (user_id, id, name, start_time, status, stop_time) VALUES (?, ?, ?, ?, ?, ?)
			`, claims["userid"], uuid, dashboard.Name, dashboard.StartTime, 1, dashboard.StopTime).WithContext(ctx).Exec()
			if err != nil {
				log.Print("Failed to update dashboard (ERROR: ", err.Error(), ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				for _, v := range dashboard.ChartData {
					uuidChartData, err := gocql.RandomUUID()
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					err = Session.Query(`
					INSERT INTO dashboard_chartdata (dashboard_id, graph_order, id, graph_color, graph_func, graph_type, metric, space_id, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
					`, uuid, v.GraphOrder, uuidChartData, v.GraphColor, v.GroupFunc, v.GraphType, v.Metric, v.SpaceID, 1).WithContext(ctx).Exec()
					if err != nil {
						log.Print("Failed to update dashboarding graphs (ERROR: ", err.Error(), ")!")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
				log.Printf("User %v successfully created dashboard!", claims["username"])
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	}
}

func putDashboardRemove(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested dashboard removal!", claims["username"])

	type Dashboard struct {
		ID string `json:"id"`
	}

	versionAPI := chi.URLParam(r, "version")
	if versionAPI == "1" {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var dashboard Dashboard
		if err := json.Unmarshal(b, &dashboard); err != nil {
			log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		{
			ctx := context.Background()
			err = Session.Query(`
			BEGIN BATCH
				DELETE FROM dashboard_chartdata WHERE dashboard_id = ?
				DELETE FROM dashboard WHERE user_id = ? AND id = ?
			APPLY BATCH
			`, dashboard.ID, claims["userid"], dashboard.ID).WithContext(ctx).Exec()
			if err != nil {
				log.Print("Failed to remove dashboard (ERROR: ", err.Error(), ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			log.Printf("User %v successfully deleted Dashboard!", claims["username"])
			w.WriteHeader(http.StatusOK)
			return
		}
	} else {
		log.Print("Failed to remove dashboard (API VERSION ERROR)!")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
