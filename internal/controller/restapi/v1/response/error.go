package response

// Error -.
type Error struct {
	Error   string            `json:"error"             example:"validation_error"`
	Message string            `json:"message"           example:"validation failed"`
	Fields  map[string]string `json:"fields,omitempty"`
} // @name v1.Error
