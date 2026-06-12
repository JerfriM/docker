# IAM Role para Lambda
resource "aws_iam_role" "lambda_role" {
  name = "api_go_lambda_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# Política básica para Lambda (logs)
resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/api-go"
  retention_in_days = 7
}

# Lambda Function
resource "aws_lambda_function" "api" {
  filename         = "../lambda.zip"
  function_name    = "api-go"
  role             = aws_iam_role.lambda_role.arn
  handler          = "bootstrap"
  runtime          = "provided.al2"
  source_code_hash = filebase64sha256("../lambda.zip")
  timeout          = 30

  environment {
    variables = {
      DATABASE_URL = var.database_url
      JWT_SECRET   = var.jwt_secret
      S3_BUCKET    = aws_s3_bucket.uploads.bucket
      SNS_TOPIC_ARN = aws_sns_topic.notifications.arn
    }
  }

  depends_on = [aws_cloudwatch_log_group.lambda_logs]
}

# API Gateway
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

# Integración Lambda con API Gateway
resource "aws_apigatewayv2_integration" "lambda" {
  api_id                 = aws_apigatewayv2_api.api.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.api.invoke_arn
  payload_format_version = "1.0"
}

# Ruta que captura todo
resource "aws_apigatewayv2_route" "default" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

# Stage de producción
resource "aws_apigatewayv2_stage" "prod" {
  api_id      = aws_apigatewayv2_api.api.id
  name        = "$default"
  auto_deploy = true
}

# Permiso para que API Gateway invoque Lambda
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*/*"
}

# Bucket S3 para archivos
resource "aws_s3_bucket" "uploads" {
  bucket = "api-go-uploads-${random_id.bucket_id.hex}"
}

resource "random_id" "bucket_id" {
  byte_length = 4
}

resource "aws_s3_bucket_public_access_block" "uploads" {
  bucket = aws_s3_bucket.uploads.id
  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "uploads_public" {
  bucket = aws_s3_bucket.uploads.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "PublicRead"
      Effect    = "Allow"
      Principal = "*"
      Action    = "s3:GetObject"
      Resource  = "${aws_s3_bucket.uploads.arn}/*"
    }]
  })
  depends_on = [aws_s3_bucket_public_access_block.uploads]
}
# SNS Topic
resource "aws_sns_topic" "notifications" {
  name = "api-go-notifications"
}

# SQS Queue
resource "aws_sqs_queue" "notifications" {
  name                       = "api-go-notifications"
  visibility_timeout_seconds = 60
}

# Suscripción SNS → SQS
resource "aws_sns_topic_subscription" "sns_to_sqs" {
  topic_arn = aws_sns_topic.notifications.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.notifications.arn
}

# Política para que SNS pueda enviar a SQS
resource "aws_sqs_queue_policy" "notifications" {
  queue_url = aws_sqs_queue.notifications.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "sns.amazonaws.com" }
      Action    = "sqs:SendMessage"
      Resource  = aws_sqs_queue.notifications.arn
      Condition = {
        ArnEquals = {
          "aws:SourceArn" = aws_sns_topic.notifications.arn
        }
      }
    }]
  })
}

# IAM Role para notification Lambda
resource "aws_iam_role" "notification_lambda_role" {
  name = "notification_lambda_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "notification_lambda_logs" {
  role       = aws_iam_role.notification_lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "notification_lambda_policy" {
  name = "notification_lambda_policy"
  role = aws_iam_role.notification_lambda_role.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["sqs:ReceiveMessage", "sqs:DeleteMessage", "sqs:GetQueueAttributes"]
        Resource = aws_sqs_queue.notifications.arn
      },
      {
        Effect   = "Allow"
        Action   = ["ses:SendEmail", "ses:SendRawEmail"]
        Resource = "*"
      }
    ]
  })
}

# Permiso SNS para Lambda principal
resource "aws_iam_role_policy" "lambda_sns_policy" {
  name = "lambda_sns_policy"
  role = aws_iam_role.lambda_role.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["sns:Publish"]
      Resource = aws_sns_topic.notifications.arn
    }]
  })
}

# CloudWatch para notification Lambda
resource "aws_cloudwatch_log_group" "notification_lambda_logs" {
  name              = "/aws/lambda/notification-lambda"
  retention_in_days = 7
}

# Notification Lambda
resource "aws_lambda_function" "notifications" {
  filename         = "../notifications.zip"
  function_name    = "notification-lambda"
  role             = aws_iam_role.notification_lambda_role.arn
  handler          = "bootstrap"
  runtime          = "provided.al2"
  source_code_hash = filebase64sha256("../notifications.zip")
  timeout          = 30

  environment {
    variables = {
      FROM_EMAIL = var.from_email
    }
  }

  depends_on = [aws_cloudwatch_log_group.notification_lambda_logs]
}

# Event Source Mapping SQS → Lambda
resource "aws_lambda_event_source_mapping" "sqs_to_lambda" {
  event_source_arn = aws_sqs_queue.notifications.arn
  function_name    = aws_lambda_function.notifications.arn
  batch_size       = 1
}

# Agregar SNS_TOPIC_ARN a Lambda principal
resource "aws_lambda_function_event_invoke_config" "api" {
  function_name = aws_lambda_function.api.function_name
}
# Permiso para que Lambda acceda a S3
resource "aws_iam_role_policy" "lambda_s3" {
  name = "lambda_s3_policy"
  role = aws_iam_role.lambda_role.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:PutObject", "s3:GetObject"]
      Resource = "${aws_s3_bucket.uploads.arn}/*"
    }]
  })
}