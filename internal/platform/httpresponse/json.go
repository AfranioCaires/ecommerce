package httpresponse

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maximumRequestBodySize = 1 << 20

type ErrorResponse struct {
	Error string `json:"error"`
}

func JSON(responseWriter http.ResponseWriter, statusCode int, payload any) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	_ = json.NewEncoder(responseWriter).Encode(payload)
}

func DecodeJSON(
	responseWriter http.ResponseWriter,
	request *http.Request,
	destination any,
) error {
	request.Body = http.MaxBytesReader(
		responseWriter,
		request.Body,
		maximumRequestBodySize,
	)

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if errorValue := decoder.Decode(destination); errorValue != nil {
		return errorValue
	}

	if errorValue := decoder.Decode(&struct{}{}); !errors.Is(errorValue, io.EOF) {
		return errors.New("the JSON request body must contain a single value")
	}

	return nil
}
