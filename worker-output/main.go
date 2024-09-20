package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/cmplx"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type FractalRequest struct {
	ID            string  `json:"id"`
	FractalType   string  `json:"fractal_type"`
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	MaxIterations int     `json:"max_iterations"`
	ColorScheme   string  `json:"color_scheme"`
	CenterX       float64 `json:"center_x,omitempty"`
	CenterY       float64 `json:"center_y,omitempty"`
	ZoomLevel     int     `json:"zoom_level,omitempty"`
}

type FractalResponse struct {
	Status         string `json:"status"`
	ImageURL       string `json:"image_url"`
	ProcessingTime int64  `json:"processing_time_ms"`
}

func generateFractal(input FractalRequest) FractalResponse {
	start := time.Now()

	img := image.NewRGBA(image.Rect(0, 0, input.Width, input.Height))
	for x := 0; x < input.Width; x++ {
		for y := 0; y < input.Height; y++ {
			real := (float64(x)/float64(input.Width))*3.5 - 2.5
			imag := (float64(y)/float64(input.Height))*2.0 - 1.0
			c := complex(real, imag)

			var colorVal int
			switch input.FractalType {
			case "Mandelbrot":
				colorVal = mandelbrot(c, input.MaxIterations)
			case "Julia":
				colorVal = julia(c, input.MaxIterations)
			case "BurningShip":
				colorVal = burningShip(c, input.MaxIterations)
			default:
				colorVal = mandelbrot(c, input.MaxIterations)
			}

			col := color.RGBA{R: uint8(colorVal * 12 % 255), G: uint8(colorVal * 7 % 255), B: uint8(colorVal * 3 % 255), A: 255}
			img.Set(x, y, col)
		}
	}

	outputFile := "data/" + input.ID + ".png"
	file, _ := os.Create(outputFile)
	defer file.Close()
	png.Encode(file, img)

	processingTime := time.Since(start).Milliseconds()
	log.Default().Printf("Generated fractal image for %s in %d ms", input.ID, processingTime)

	return FractalResponse{
		Status:         "completed",
		ImageURL:       fmt.Sprintf("https://example.com/fractal_images/%s", outputFile),
		ProcessingTime: processingTime,
	}
}

func mandelbrot(c complex128, maxIter int) int {
	z := complex(0, 0)
	for n := 0; n < maxIter; n++ {
		if cmplx.Abs(z) > 2 {
			return n
		}
		z = z*z + c
	}
	return maxIter
}

func julia(c complex128, maxIter int) int {
	constant := complex(-0.7, 0.27015)
	z := c
	for n := 0; n < maxIter; n++ {
		if cmplx.Abs(z) > 2 {
			return n
		}
		z = z*z + constant
	}
	return maxIter
}

func burningShip(c complex128, maxIter int) int {
	z := complex(0, 0)
	for n := 0; n < maxIter; n++ {
		if cmplx.Abs(z) > 2 {
			return n
		}
		z = complex(abs(real(z)), abs(imag(z)))*complex(abs(real(z)), abs(imag(z))) + c
	}
	return maxIter
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func processSQSMessages(queueURL string) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Error loading AWS configuration:", err)
		return
	}

	client := sqs.NewFromConfig(cfg)

	resp, err := client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     10,
	})
	if err != nil {
		fmt.Println("Error receiving message from SQS:", err)
		return
	}

	if len(resp.Messages) == 0 {
		fmt.Println("No messages received")
		return
	}

	for _, message := range resp.Messages {
		var fractalRequest FractalRequest
		err := json.Unmarshal([]byte(*message.Body), &fractalRequest)
		if err != nil {
			fmt.Println("Error unmarshalling message body:", err)
			continue
		}

		output := generateFractal(fractalRequest)
		result, _ := json.Marshal(output)
		fmt.Println(string(result))

		_, err = client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
			QueueUrl:      &queueURL,
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			fmt.Println("Error deleting message from SQS:", err)
		} else {
			fmt.Println("Message processed and deleted from SQS.")
		}
	}
}

func main() {
	outputQueueURL := os.Getenv("OUTPUT_QUEQUE_URL")

	for {
		processSQSMessages(outputQueueURL)
	}
}
