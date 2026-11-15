output "weather_url" {
  description = "GET /weather URL"
  value       = "${aws_apigatewayv2_api.api.api_endpoint}/weather"
}
