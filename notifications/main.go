package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type NotificationMessage struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Source  string `json:"source,omitempty"` // Para identificar si viene de EventBridge
}

// Cambiamos la firma para recibir json.RawMessage (bytes puros)
func handler(ctx context.Context, rawEvent json.RawMessage) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error configurando AWS: %v", err)
		return err
	}

	sesClient := ses.NewFromConfig(cfg)
	fromEmail := os.Getenv("FROM_EMAIL")

	// 1. Intentar parsear como un evento directo de EventBridge Scheduler
	var directMsg NotificationMessage
	if err := json.Unmarshal(rawEvent, &directMsg); err == nil && directMsg.Source == "eventbridge-scheduler" {
		log.Printf("--- Detectado evento directo de EventBridge Scheduler ---")
		return sendEmail(ctx, sesClient, fromEmail, directMsg)
	}

	// 2. Si no es de EventBridge, lo procesamos con la estructura normal de SQS
	log.Printf("--- Detectado evento SQS/SNS ---")
	var sqsEvent events.SQSEvent
	if err := json.Unmarshal(rawEvent, &sqsEvent); err != nil {
		log.Printf("Error: El evento no coincide con SQS ni con EventBridge: %v", err)
		return err
	}

	for _, record := range sqsEvent.Records {
		log.Printf("Procesando mensaje SQS: %s", record.Body)

		var snsWrapper struct {
			Message string `json:"Message"`
		}
		if err := json.Unmarshal([]byte(record.Body), &snsWrapper); err != nil {
			log.Printf("Error parseando SNS wrapper: %v", err)
			continue
		}

		var msg NotificationMessage
		if err := json.Unmarshal([]byte(snsWrapper.Message), &msg); err != nil {
			log.Printf("Error parseando mensaje interno de SNS: %v", err)
			continue
		}

		if err := sendEmail(ctx, sesClient, fromEmail, msg); err != nil {
			return err
		}
	}

	return nil
}

// Función auxiliar para evitar duplicar la lógica de envío de SES
func sendEmail(ctx context.Context, client *ses.Client, from string, msg NotificationMessage) error {
	log.Printf("Enviando email a: %s", msg.Email)

	_, err := client.SendEmail(ctx, &ses.SendEmailInput{
		Source: &from,
		Destination: &types.Destination{
			ToAddresses: []string{msg.Email},
		},
		Message: &types.Message{
			Subject: &types.Content{Data: &msg.Subject},
			Body: &types.Body{
				Text: &types.Content{Data: &msg.Message},
			},
		},
	})
	if err != nil {
		log.Printf("Error enviando email a %s: %v", msg.Email, err)
		return err
	}

	log.Printf("Email enviado exitosamente a: %s", msg.Email)
	return nil
}

func main() {
	lambda.Start(handler)
}
