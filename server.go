package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"github.com/gorilla/mux"
)

type TimingsResponse struct {
	IsError bool
	ErrorMsg string
	NumStreamsAnalyzed int
	BinFreqs []int64
	BinTimes []float64
	DowBins []int64
	Username string
	Durations []float64
}

type TimingsRequest struct {
	Username string
}

func error_response(w http.ResponseWriter, reason string, status int) {
	var resp TimingsResponse
	resp.IsError = true
	resp.ErrorMsg = reason
	resp.NumStreamsAnalyzed = 0
	resp.BinFreqs = nil
	resp.BinTimes = nil

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(resp)
	fmt.Println("Sending error")
	if err != nil {
		log.Println(err)
	}
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got request", r.Method, r.URL)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == "OPTIONS" {
		_, err := w.Write(nil);
		if err != nil {
			log.Println(err);
		}
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	fmt.Println("Got body", body)
	if err != nil {
		log.Println(err)
		error_response(w, "Could not read request body", http.StatusInternalServerError)
		return
	}
	var req TimingsRequest
	err = json.Unmarshal(body, &req);
	if err != nil {
		log.Println(err)
		error_response(w, "Could not parse json", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		log.Println("Request was missing username")
		error_response(w, "Missing username", http.StatusBadRequest)
		return
	}

	n, sum, timings, dow_sum, durs := get_bin_sum_and_timings(req.Username)

	if n <= 0 {
		error_response(w, "Found no vods", http.StatusOK)
		return
	}

	var resp TimingsResponse
	resp.IsError = false
	resp.ErrorMsg = ""
	resp.NumStreamsAnalyzed = n
	resp.BinTimes = timings
	resp.BinFreqs = sum
	resp.Username = req.Username
	resp.DowBins = dow_sum
	resp.Durations = durs


	w.Header().Set("Content-Type", "application/json")
	fmt.Println("Sending successful response")
	err = json.NewEncoder(w).Encode(resp)

	if err != nil {
		log.Println(err)
	}
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/timings", homeLink).Methods("POST", "OPTIONS")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./srv/")))
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("No PORT found")
	}
	port_str := fmt.Sprintf(":%s", port)
	fmt.Println("Now serving on", port_str)
	log.Fatal(http.ListenAndServe(port_str, router))
}