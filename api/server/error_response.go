package server

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"net/http"
)

var ErrInternalServerError = errors.New("Something unexpected happened on the server")

func simpleError(err error) *models.Error {
	return &models.Error{Error: &models.ErrorBody{Message: err.Error()}}
}

var errStatusCode = map[error]int{
	models.ErrAppsNotFound:        http.StatusNotFound,
	models.ErrAppsAlreadyExists:   http.StatusConflict,
	models.ErrRoutesNotFound:      http.StatusNotFound,
	models.ErrRoutesAlreadyExists: http.StatusConflict,
}

func handleErrorResponse(c *gin.Context, r RequestController, err error) {
	log := r.Logger()
	log.Error(err)

	if code, ok := errStatusCode[err]; ok {
		c.JSON(code, simpleError(err))
	} else {
		c.JSON(http.StatusInternalServerError, simpleError(err))
	}
}
