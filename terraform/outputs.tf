output "api_gateway_url" {
  description = "URL publica de la API"
  value       = aws_apigatewayv2_api.api.api_endpoint
}

output "lambda_function_name" {
  value = aws_lambda_function.api.function_name
}

output "sns_topic_arn" {
  value = aws_sns_topic.notifications.arn
}

output "sqs_queue_url" {
  value = aws_sqs_queue.notifications.url
}

output "notification_lambda_name" {
  value = aws_lambda_function.notifications.function_name
}

output "eventbridge_schedule_name" {
  value = aws_scheduler_schedule.every_5_minutes.name
}