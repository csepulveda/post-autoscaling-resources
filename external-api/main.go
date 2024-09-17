package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Fractal struct {
	ID            string  `json:"id"`
	FractalType   string  `json:"fractal_type"`
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	MaxIterations int     `json:"max_iterations"`
	ColorScheme   string  `json:"color_scheme"`
	CenterX       float64 `json:"center_x"`
	CenterY       float64 `json:"center_y"`
	ZoomLevel     int     `json:"zoom_level"`
}

func RandomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func RandomFractal(id string) Fractal {
	fractalTypes := []string{"Mandelbrot", "Julia", "BurningShip"}
	colorSchemes := []string{"rainbow", "fire", "blue_shades", "monochrome", "pastel"}

	return Fractal{
		ID:            id,
		FractalType:   fractalTypes[rand.Intn(len(fractalTypes))],
		Width:         rand.Intn(1920-800+1) + 800,
		Height:        rand.Intn(1080-600+1) + 600,
		MaxIterations: rand.Intn(2000-500+1) + 500,
		ColorScheme:   colorSchemes[rand.Intn(len(colorSchemes))],
		CenterX:       RandomFloat(-2.0, 2.0),
		CenterY:       RandomFloat(-2.0, 2.0),
		ZoomLevel:     rand.Intn(50) + 1,
	}
}

func fractalHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 || pathParts[2] == "" {
		http.Error(w, "Fractal ID not provided", http.StatusBadRequest)
		return
	}
	id := pathParts[2]

	time.Sleep(time.Duration(rand.Intn(300)+1500) * time.Millisecond)
	fractal := RandomFractal(id)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fractal); err != nil {
		http.Error(w, "Failed to generate fractal", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/fractals/", fractalHandler)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
