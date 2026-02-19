package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTerms returns all known terms.
func (h *Handler) GetTerms(c *gin.Context) {
	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	terms, hasNext, err := h.db.QueryAllTerms(c.Request.Context(), params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination := buildPaginationMeta(params, len(terms), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(terms),
		"terms":      terms,
		"pagination": pagination,
	})
}
