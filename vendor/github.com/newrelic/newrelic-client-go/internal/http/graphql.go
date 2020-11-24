package http

import (
	"encoding/json"
	"strings"
)

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data interface{} `json:"data"`
}

// GraphQLError represents a single error.
type GraphQLError struct {
	Message    string   `json:"message,omitempty"`
	Path       []string `json:"path,omitempty"`
	Extensions struct {
		ErrorClass string `json:"errorClass,omitempty"`
	} `json:"extensions,omitempty"`
	DownstreamResponse []GraphQLDownstreamResponse `json:"downstreamResponse,omitempty"`
}

// GraphQLDownstreamResponse represents an error's downstream response.
type GraphQLDownstreamResponse struct {
	Extensions struct {
		Code             string `json:"code,omitempty"`
		ValidationErrors []struct {
			Name   string `json:"name,omitempty"`
			Reason string `json:"reason,omitempty"`
		} `json:"validationErrors,omitempty"`
	} `json:"extensions,omitempty"`
	Message string `json:"message,omitempty"`
}

// GraphQLErrorResponse represents a default error response body.
type GraphQLErrorResponse struct {
	Errors []GraphQLError `json:"errors"`
}

func (r *GraphQLErrorResponse) Error() string {
	if len(r.Errors) > 0 {
		messages := []string{}
		for _, e := range r.Errors {

			if e.Message != "" {
				messages = append(messages, e.Message)
			}

			if e.DownstreamResponse != nil {
				f, _ := json.Marshal(e.DownstreamResponse)
				messages = append(messages, string(f))
			}
		}
		return strings.Join(messages, ", ")
	}

	return ""
}

// IsNotFound determines if the error is due to a missing resource.
func (r *GraphQLErrorResponse) IsNotFound() bool {
	return false
}

// IsRetryableError determines if the error is due to a server timeout, or another error that we might want to retry.
func (r *GraphQLErrorResponse) IsRetryableError() bool {
	if len(r.Errors) == 0 {
		return false
	}

	for _, err := range r.Errors {
		if err.Extensions.ErrorClass == "TIMEOUT" {
			return true
		}

		if err.Extensions.ErrorClass == "INTERNAL_SERVER_ERROR" {
			return true
		}

		for _, downstreamErr := range err.DownstreamResponse {
			if downstreamErr.Extensions.Code == "INTERNAL_SERVER_ERROR" {
				return true
			}
		}
	}

	return false
}

// New creates a new instance of GraphQLErrorRepsonse.
func (r *GraphQLErrorResponse) New() ErrorResponse {
	return &GraphQLErrorResponse{}
}
