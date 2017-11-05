package bitgo

// Error is the response returned when a call is unsuccessful.
type Error struct {
	HTTPStatusCode int
	Message        string `json:"error"`
}

func (e Error) Error() string {
	return e.Message
}
