package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
)

const port = ":8888"

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/scan", cors(handleScan))
	mux.HandleFunc("/images", cors(handleImages))

	log.Println("eagle-scout backend listening on", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func handleImages(w http.ResponseWriter, r *http.Request) {
	out, err := exec.Command("docker", "images",
		"--format", `{"repository":"{{.Repository}}","tag":"{{.Tag}}","id":"{{.ID}}","size":"{{.Size}}","created":"{{.CreatedSince}}"}`).Output()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var images []json.RawMessage
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			images = append(images, json.RawMessage(line))
		}
	}
	if images == nil {
		images = []json.RawMessage{}
	}

	writeJSON(w, images)
}

func handleScan(w http.ResponseWriter, r *http.Request) {
	image := r.URL.Query().Get("image")
	if image == "" {
		http.Error(w, "image parameter required", http.StatusBadRequest)
		return
	}

	type result struct {
		output string
		err    error
	}

	var (
		wg              sync.WaitGroup
		qv, cves, recs result
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		out, err := exec.Command("docker", "scout", "quickview", image).CombinedOutput()
		qv = result{string(out), err}
	}()
	go func() {
		defer wg.Done()
		out, err := exec.Command("docker", "scout", "cves", image).CombinedOutput()
		cves = result{string(out), err}
	}()
	go func() {
		defer wg.Done()
		out, err := exec.Command("docker", "scout", "recommendations", image).CombinedOutput()
		recs = result{string(out), err}
	}()
	wg.Wait()

	writeJSON(w, map[string]string{
		"image":           image,
		"quickview":       qv.output,
		"cves":            cves.output,
		"recommendations": recs.output,
		"error":           errStr(qv.err),
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func errStr(err error) string {
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return ""
}
