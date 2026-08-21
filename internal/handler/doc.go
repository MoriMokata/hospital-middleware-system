// Package handler contains the Gin HTTP handlers for each endpoint
// (/staff/create, /staff/login, /patient/search, /health). Handlers stay
// thin: parse/validate input, call a service, map the result to a response.
package handler
