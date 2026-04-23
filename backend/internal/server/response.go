package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// successResponse is the envelope for all successful API responses.
type successResponse struct {
	Data any `json:"data"`
}

// errorResponse is the envelope for all error API responses.
type errorResponse struct {
	Error string `json:"error"`
}

// paginatedResponse wraps paginated list results.
type paginatedResponse struct {
	Data     any   `json:"data"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// respondOK sends HTTP 200 with a JSON success envelope.
func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, successResponse{Data: data})
}

// respondCreated sends HTTP 201 with a JSON success envelope.
func respondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, successResponse{Data: data})
}

// respondPaginated sends HTTP 200 with pagination metadata.
func respondPaginated(c *gin.Context, data any, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, paginatedResponse{
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// respondError sends an HTTP error with a JSON error envelope.
func respondError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, errorResponse{Error: msg})
}

// respondBadRequest is a shortcut for 400.
func respondBadRequest(c *gin.Context, msg string) {
	respondError(c, http.StatusBadRequest, msg)
}

// respondUnauthorized is a shortcut for 401.
func respondUnauthorized(c *gin.Context, msg string) {
	respondError(c, http.StatusUnauthorized, msg)
}

// respondForbidden is a shortcut for 403.
func respondForbidden(c *gin.Context, msg string) {
	respondError(c, http.StatusForbidden, msg)
}

// respondNotFound is a shortcut for 404.
func respondNotFound(c *gin.Context, msg string) {
	respondError(c, http.StatusNotFound, msg)
}

// respondInternalError is a shortcut for 500.
func respondInternalError(c *gin.Context, msg string) {
	respondError(c, http.StatusInternalServerError, msg)
}
