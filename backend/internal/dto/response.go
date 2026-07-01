package dto

// Response is the standard envelope returned by most API endpoints.
type Response struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Data    any               `json:"data"`
	Errors  map[string]string `json:"errors"`
}
