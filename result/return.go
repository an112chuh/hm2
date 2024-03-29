package result

import (
	"encoding/json"
	"hm2/report"
	"net/http"
)

type Returning interface{}

func ReturnJSON(w http.ResponseWriter, object Returning) {
	ansB, err := json.Marshal(object)
	if err != nil {
		report.ErrorServer(nil, err)
	}
	Headers(w)
	_, err = w.Write(ansB)
	if err != nil {
		report.ErrorServer(nil, err)
	}
}

func Headers(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,PUT,PATCH,POST,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "content-type")
}
