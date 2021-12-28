package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	lru "github.com/hashicorp/golang-lru"
	nanoid "github.com/matoous/go-nanoid"
)

var (
	mem = func() *lru.Cache {
		c, _ := lru.New(1024)
		return c
	}()
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/flag", handlePostFlag).Methods("POST")
	r.HandleFunc("/flag/{id}", handleGetFlag).Methods("GET")
	r.HandleFunc("/last", handleLast).Methods("GET")
	r.HandleFunc("/", handleIndex).Methods("GET")

	s := &http.Server{
		Handler:      r,
		Addr:         ":2001",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Starting to listen at %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./main.html")
}

func handleGetFlag(w http.ResponseWriter, r *http.Request) {
	flagID := mux.Vars(r)["id"]
	if flagID == "" {
		reportError(w, http.StatusBadRequest, fmt.Errorf("id not found"))
		return
	}

	flag, ok := mem.Get(flagID)
	if !ok {
		reportError(w, http.StatusNotFound, fmt.Errorf("flag not found"))
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"flag":   flag,
	})
}

func handlePostFlag(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	var flagDTO struct {
		Flag string `json:"flag"`
	}
	if err := json.NewDecoder(r.Body).Decode(&flagDTO); err != nil {
		reportError(w, http.StatusBadRequest, err)
		return
	}

	id, err := nanoid.Nanoid()
	if err != nil {
		reportError(w, http.StatusInternalServerError, err)
		return
	}

	mem.Add(id, struct {
		Flag  string    `json:"flag"`
		Added time.Time `json:"added"`
	}{
		Flag:  flagDTO.Flag,
		Added: time.Now(),
	})

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"id":     id,
	})
}

func handleLast(w http.ResponseWriter, r *http.Request) {
	n := 10
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if newN, err := strconv.Atoi(limit); err != nil {
			reportError(w, http.StatusBadRequest, fmt.Errorf("incorrect limit: %s", err))
			return
		} else if newN > 0 && newN < 100 {
			n = newN
		}
	}

	keys := mem.Keys()
	if n > len(keys) {
		n = len(keys)
	}

	flags := make([]interface{}, n)
	for i := 0; i < n; i++ {
		flags[i] = keys[len(keys) - n + i]
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"last":  flags,
	})
}

func reportError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "err",
		"error":  err.Error(),
	})
}
