output "api_gateway_url" {
  description = "URL publica de la API"
  value       = aws_apigatewayv2_api.api.api_endpoint
}

output "lambda_function_name" {
  value = aws_lambda_function.api.function_name
}