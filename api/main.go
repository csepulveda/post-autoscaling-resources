package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"golang.org/x/exp/rand"
)

type PostRequest struct {
	ID string `json:"id"`
}

var (
	sqsClient *sqs.Client
	queueURL  string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Error loading AWS configuration: %v", err)
	}

	sqsClient = sqs.NewFromConfig(cfg)
	queueURL = os.Getenv("QUEQUE_URL")
}

func handleSend(w http.ResponseWriter, r *http.Request) {
	var postRequest PostRequest
	err := json.NewDecoder(r.Body).Decode(&postRequest)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = doSomeProcess(postRequest)
	if err != nil {
		http.Error(w, "Failed to do some process", http.StatusInternalServerError)
		return
	}

	messageBody, err := json.Marshal(postRequest)
	if err != nil {
		http.Error(w, "Failed to create message body", http.StatusInternalServerError)
		return
	}

	_, err = sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: aws.String(string(messageBody)),
	})
	if err != nil {
		http.Error(w, "Failed to send message to SQS", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Message sent and ID stored successfully.")
}

func doSomeProcess(PostRequest PostRequest) error {
	// wait randmly between 200 and 400ms
	time.Sleep(time.Duration(rand.Intn(200)+200) * time.Millisecond)
	log.Default().Printf("Processed ID: %s", PostRequest.ID)
	return nil
}

func main() {
	http.HandleFunc("/send", handleSend)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	port := "8080"
	fmt.Println("Server is running on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
