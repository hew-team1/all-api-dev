package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/connpass", HackathonGet).Methods("GET")
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodPut,
			http.MethodDelete,
		},
	}).Handler(r)

	fmt.Println("サーバー起動 : 60003 port で受信")

	// log.Fatal は、異常を検知すると処理の実行を止めてくれる
	log.Fatal(http.ListenAndServe(":60003", c))
}

// ==================== GET ====================
type ConnpassApi struct {
	ResultsReturned int `json:"results_returned"`
	Events          []struct {
		EventURL      string `json:"event_url"`
		EventType     string `json:"event_type"`
		OwnerNickname string `json:"owner_nickname"`
		Series        struct {
			URL   string `json:"url"`
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"series"`
		UpdatedAt        time.Time `json:"updated_at"`
		Lat              string    `json:"lat"`
		StartedAt        time.Time `json:"started_at"`
		HashTag          string    `json:"hash_tag"`
		Title            string    `json:"title"`
		EventID          int       `json:"event_id"`
		Lon              string    `json:"lon"`
		Waiting          int       `json:"waiting"`
		Limit            int       `json:"limit"`
		OwnerID          int       `json:"owner_id"`
		OwnerDisplayName string    `json:"owner_display_name"`
		Description      string    `json:"description"`
		Address          string    `json:"address"`
		Catch            string    `json:"catch"`
		Accepted         int       `json:"accepted"`
		EndedAt          time.Time `json:"ended_at"`
		Place            string    `json:"place"`
	} `json:"events"`
	ResultsStart     int `json:"results_start"`
	ResultsAvailable int `json:"results_available"`
}
type HackathonResponse struct {
	EventID   int    `json:"event_id"`
	EventURL  string `json:"event_url"`
	Title     string `json:"title"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
}

func HackathonGet(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	addNow := now.AddDate(0, 1, 0)
	formatNow := now.Format("200601")
	formatAddNow := addNow.Format("200601")

	url := "https://connpass.com/api/v1/event/?keyword_or=ハッカソン&keyword_or=hackathon&keyword_or=hack&count=100&order=2&ym=" + formatNow + "&ym=" + formatAddNow
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)

	jsonBytes := byteArray
	var data ConnpassApi

	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return
	}
	var resHackathon []HackathonResponse
	for _, row := range data.Events {
		if row.StartedAt.Before(now) {
			break
		}
		resHackathon = append(
			resHackathon,
			HackathonResponse{
				EventID:   row.EventID,
				EventURL:  row.EventURL,
				Title:     row.Title,
				StartedAt: row.StartedAt.Format("2006/01/02"),
				EndedAt:   row.EndedAt.Format("2006/01/02"),
			},
		)
	}

	j, _ := json.Marshal(resHackathon)
	w.Write(j)

	// 取得値のログ
	fmt.Println(string(j))
}
