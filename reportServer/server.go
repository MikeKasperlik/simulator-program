package reportServer

import (
	"encoding/json"
	"net/http"
)

type ReportData struct {
	Speed float64 `json:"speed"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
}

type DataChan chan ReportData

type ReportServer struct {
	IncomingData DataChan
}

func New(incomingData DataChan) *ReportServer {
	return &ReportServer{
		IncomingData: incomingData,
	}
}

func (s *ReportServer) handleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Use method POST", http.StatusNotFound)
		return
	}

	var data ReportData
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)

	if err != nil {
		http.Error(w, "JSON encoding error", http.StatusInternalServerError)
		return
	}

	s.IncomingData <- data

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte("OK\n"))
}

func (s *ReportServer) Start() {
	http.HandleFunc("/report", s.handleReport)
	http.ListenAndServe(":8080", nil)
}
