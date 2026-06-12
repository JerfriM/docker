variable "aws_region" {
  default = "us-east-1"
}

variable "database_url" {
  description = "URL de conexion a Neon PostgreSQL"
  sensitive   = true
}
variable "from_email" {
  description = "Email verificado en SES para enviar notificaciones"
}

variable "jwt_secret" {
  description = "Clave secreta para JWT"
  sensitive   = true
}