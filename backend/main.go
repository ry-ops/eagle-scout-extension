package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const socketPath = "/run/guest-services/eagle-scout.sock"

func main() {
	os.Remove(socketPath)
	os.MkdirAll("/run/guest-services", 0755)

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("failed to listen on socket: %v", err)
	}
	defer ln.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/quickview", handleQuickview)
	mux.HandleFunc("/cves", handleCVEs)
	mux.HandleFunc("/recommendations", handleRecommendations)
	mux.HandleFunc("/images", handleImages)

	log.Println("eagle-scout backend listening on", socketPath)
	log.Fatal(http.Serve(ln, mux))
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

func handleQuickview(w http.ResponseWriter, r *http.Request) {
	image := r.URL.Query().Get("image")
	if image == "" {
		http.Error(w, "image parameter required", http.StatusBadRequest)
		return
	}

	out, err := exec.Command("docker", "scout", "quickview", image).CombinedOutput()
	writeJSON(w, map[string]string{
		"image":  image,
		"output": string(out),
		"error":  errStr(err),
	})
}

func handleCVEs(w http.ResponseWriter, r *http.Request) {
	image := r.URL.Query().Get("image")
	if image == "" {
		http.Error(w, "image parameter required", http.StatusBadRequest)
		return
	}

	args := []string{"scout", "cves", image}
	if r.URL.Query().Get("only_fixed") == "true" {
		args = append(args, "--only-fixed")
	}
	if sev := r.URL.Query().Get("severity"); sev != "" {
		args = append(args, "--only-severity", sev)
	}

	out, err := exec.Command("docker", args...).CombinedOutput()
	writeJSON(w, map[string]string{
		"image":  image,
		"output": string(out),
		"error":  errStr(err),
	})
}

func handleRecommendations(w http.ResponseWriter, r *http.Request) {
	image := r.URL.Query().Get("image")
	if image == "" {
		http.Error(w, "image parameter required", http.StatusBadRequest)
		return
	}

	out, err := exec.Command("docker", "scout", "recommendations", image).CombinedOutput()
	writeJSON(w, map[string]string{
		"image":  image,
		"output": string(out),
		"error":  errStr(err),
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
