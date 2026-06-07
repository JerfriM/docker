resource "aws_apigatewayv2_api" "api" {
  name          = "api-go-gateway"
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins     = ["*"]
    allow_methods     = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers     = ["*"]
    expose_headers    = ["*"]
    max_age           = 86400
    allow_credentials = false
  }
}