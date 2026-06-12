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
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error configurando AWS: %v", err)
		return err
	}

	sesClient := ses.NewFromConfig(cfg)
	fromEmail := os.Getenv("FROM_EMAIL")

	for _, record := range sqsEvent.Records {
		log.Printf("Procesando mensaje: %s", record.Body)

		var snsWrapper struct {
			Message string `json:"Message"`
		}
		if err := json.Unmarshal([]byte(record.Body), &snsWrapper); err != nil {
			log.Printf("Error parseando SNS wrapper: %v", err)
			continue
		}

		var msg NotificationMessage
		if err := json.Unmarshal([]byte(snsWrapper.Message), &msg); err != nil {
			log.Printf("Error parseando mensaje: %v", err)
			continue
		}

		log.Printf("Enviando email a: %s", msg.Email)

		_, err = sesClient.SendEmail(ctx, &ses.SendEmailInput{
			Source: &fromEmail,
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
			log.Printf("Error enviando email: %v", err)
			return err
		}
		log.Printf("Email enviado exitosamente a: %s", msg.Email)
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
