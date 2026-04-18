package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/DemianSV/chrdsclient"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gocql/gocql"
)

func putChartData(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	log.Printf("User %v requested chart data", claims["username"])

	type RequestT struct {
		DataSRC string `json:"datasrc"`
	}

	type ChartDataSetT struct {
		BackgroundColor string    `json:"backgroundColor"`
		Data            []float32 `json:"data"`
		BarPercentage   float32   `json:"barPercentage"` // Пространство между столбиками
	}

	type ChartDataT struct {
		Labels   []int64         `json:"labels"`
		DataSets []ChartDataSetT `json:"datasets"`
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

			var request RequestT
			if err := json.Unmarshal(b, &request); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var chartDataSet ChartDataSetT
			var chartData ChartDataT
			ctx := context.Background()
			var scanner gocql.Scanner
			stopTime := time.Now().UnixMilli()
			startTime := stopTime - (23 * 60 * 60 * 1000)

			chartDataSet.BackgroundColor = "#ade2ffbe"
			chartDataSet.BarPercentage = 1.1 // Пространство между столбиками

			var spaceID []string

			{
				switch claims["role"] {
				case "superadmin":
					scanner = Session.Query(`SELECT id, status FROM space`).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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
						scanner = Session.Query(`SELECT id, status FROM space WHERE id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					}
				default:
					var spaceList []string
					scanner = Session.Query(`SELECT space_id FROM user_space WHERE user_id = ?`, claims["ownerid"]).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
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
						scanner = Session.Query(`SELECT id, status FROM space WHERE id IN (?`+strings.Repeat(", ?", len(spaceList)-1)+`)`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					}
				}

				for scanner.Next() {
					var id string
					var status int
					err := scanner.Scan(&id, &status)
					if err != nil {
						go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					if status == 1 {
						spaceID = append(spaceID, id)
					}
				}
			}

			if len(spaceID) > 0 {
				for iTime := startTime; iTime <= stopTime; iTime = iTime + (60 * 60 * 1000) {
					chartData.Labels = append(chartData.Labels, iTime)

					args := []interface{}{}
					for _, item := range spaceID {
						args = append(args, item)
					}
					args = append(args, iTime)
					args = append(args, iTime+(60*60*1000))

					var scanner gocql.Scanner
					switch request.DataSRC {
					case "raw_text":
						scanner = Session.Query(`SELECT sum(value) FROM raw_count WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceID)-1)+`) AND raw_type = 'text' AND create_time >= ? AND create_time < ?`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					case "raw_data":
						scanner = Session.Query(`SELECT sum(value) FROM raw_count WHERE space_id IN (?`+strings.Repeat(", ?", len(spaceID)-1)+`) AND raw_type = 'data' AND create_time >= ? AND create_time < ?`, args...).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
					default:
						go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					for scanner.Next() {
						var count int64
						err := scanner.Scan(&count)
						if err != nil {
							go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
							w.WriteHeader(http.StatusInternalServerError)
							return
						} else {
							chartDataSet.Data = append(chartDataSet.Data, float32(count))
						}
					}
					if err := scanner.Err(); err != nil {
						log.Print(err)
						go chrdsclient.Metric("httpstatus", float32(http.StatusInternalServerError))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			}

			chartData.DataSets = append(chartData.DataSets, chartDataSet)

			responseJSON, err := json.Marshal(chartData)
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
