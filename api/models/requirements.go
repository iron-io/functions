package models

import (
	apiErrors "github.com/go-openapi/errors"
)

type Requirements struct {
	Memory string `json:"req_memory,omitempty"`
}

func (r *Requirements) Validate() error {
	var res []error

	if r.Memory == "" {
		r.Memory = "32M"
	}

	if len(res) > 0 {
		return apiErrors.CompositeValidationError(res...)
	}

	return nil
}
