package main

import (
	"net/http"
	"encoding/json"
	"io"
)

func test_handler(w http.ResponseWriter, r *http.Request, args []string, pd []byte) {
	post_data := pd
	log.Info("post data", string(post_data))

	data := map[string]interface{}{
		"err": 0,

	}
	res, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, string(res))
}