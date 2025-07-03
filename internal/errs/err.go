package errs

import (
	"errors"
	"fmt"
	"github.com/aws/smithy-go"
	"net/http"
	"strings"
)

type ApiError struct {
	StatusCode  int
	FullError   error
	UserMessage string
}

func NewApiError(statusCode int, err error, userMessage string) *ApiError {
	return &ApiError{
		StatusCode:  statusCode,
		FullError:   err,
		UserMessage: userMessage,
	}
}

func (e *ApiError) Error() string {
	var out []string
	if e.UserMessage != "" {
		out = append(out, e.UserMessage)
	}
	if e.FullError != nil {
		out = append(out, e.FullError.Error())
	}
	if e.StatusCode != 0 {
		out = append(out, http.StatusText(e.StatusCode))
	}
	return strings.Join(out, ": ")
}

// FromAwsApi wrap AWS SDK error in API error
func FromAwsApi(err error, msg string) error {
	var apiErr smithy.APIError
	if ok := errors.As(err, &apiErr); ok {
		code := apiErr.ErrorCode()
		if strings.HasSuffix(code, ".NotFound") {
			return NewApiError(http.StatusNotFound, err, fmt.Sprintf("%s: not found", msg))
		}
		if strings.HasSuffix(code, ".Duplicate") {
			return NewApiError(http.StatusConflict, err, fmt.Sprintf("%s: conflict", msg))
		}
		return NewApiError(http.StatusInternalServerError, err, fmt.Sprintf("%s: internal server error", msg))
	}
	return NewApiError(0, err, msg)
}
