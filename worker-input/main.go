package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
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

type PostRequest struct {
	ID string `json:"id"`
}

func fetchFractalRequest(id string) (FractalRequest, error) {
	url := fmt.Sprintf("%s/fractals/%s", os.Getenv("FRACTAL_API_BASE_URL"), id)
	resp, err := http.Get(url)
	if err != nil {
		return FractalRequest{}, err
	}
	defer resp.Body.Close()

	var fractalRequest FractalRequest
	err = json.NewDecoder(resp.Body).Decode(&fractalRequest)
	if err != nil {
		return FractalRequest{}, err
	}

	return fractalRequest, nil
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
		var postRequest PostRequest
		err := json.Unmarshal([]byte(*message.Body), &postRequest)
		if err != nil {
			fmt.Println("Error unmarshalling message body:", err)
			continue
		}

		fractalRequest, err := fetchFractalRequest(postRequest.ID)
		if err != nil {
			fmt.Println("Error fetching fractal request from external API:", err)
			continue
		}

		err = sendMessageFractal(fractalRequest, client)
		if err != nil {
			fmt.Println("Error sending message to output queue:", err)
			continue
		}

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

func sendMessageFractal(fractalRequest FractalRequest, client *sqs.Client) error {
	outputQueueURL := os.Getenv("OUTPUT_QUEQUE_URL")
	messageBody, err := json.Marshal(fractalRequest)
	if err != nil {
		return err
	}

	_, err = client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    &outputQueueURL,
		MessageBody: aws.String(string(messageBody)),
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	inputQueueURL := os.Getenv("INPUT_QUEQUE_URL")

	for {
		processSQSMessages(inputQueueURL)
	}
}
