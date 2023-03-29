package main

import (
	"fmt"
	"github.com/edgenesis/shifu/pkg/logger"
	"net/http"
	"os"
	"time"
)

var image_counter = 0

func main() {
	http.HandleFunc("/capture", captureHandler)
	http.HandleFunc("/recognize/image", recognitionHandler)
	http.HandleFunc("/status", healthHandler)

	port := "8080"
	logger.Infof("Starting server on port %s...", port)
	logger.Fatal(http.ListenAndServe(":"+port, nil))
}

func captureHandler(w http.ResponseWriter, r *http.Request) {
	imagePath := fmt.Sprintf("images/image%d.png", image_counter)
	image_counter = (image_counter + 1) % 5
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		logger.Error(err)
		http.Error(w, "Error reading image file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(imageData)
}

func recognitionHandler(w http.ResponseWriter, r *http.Request) {
	imagePath := fmt.Sprintf("images/recognize%d.png", image_counter)
	image_counter = (image_counter + 1) % 5
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		logger.Error(err)
		http.Error(w, "Error reading image file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(imageData)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"timestamp":"%d","status":"active"}`, time.Now().UnixMilli())))
}
