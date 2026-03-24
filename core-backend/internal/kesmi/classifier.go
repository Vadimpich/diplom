package kesmi

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"diplom/internal/decision"
)

type FailureClassification struct {
	State      string
	ErrorClass string
	Retryable  bool
}

func ClassifyFailure(err error, httpStatus int, errorCode string) FailureClassification {
	errorCode = strings.ToLower(strings.TrimSpace(errorCode))
	if isTimeoutOrNetwork(err) || httpStatus >= 500 || errorCode == "pool_busy" || errorCode == "pool_exhausted" {
		return FailureClassification{
			State:      decision.DecisionStateTransportExhausted,
			ErrorClass: "transport",
			Retryable:  true,
		}
	}
	return FailureClassification{
		State:      decision.DecisionStateBusinessError,
		ErrorClass: "business",
		Retryable:  false,
	}
}

func isTimeoutOrNetwork(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "timeout") || strings.Contains(message, "connection refused") || strings.Contains(message, "no route to host")
}
