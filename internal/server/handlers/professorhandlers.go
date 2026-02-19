package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProfessorByID loads a professor by ID.
func (h *Handler) GetProfessorByID(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Professor ID is required"})
		return
	}

	professor, err := h.db.GetProfessorById(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get professor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"professor": professor,
	})
}

// GetProfessorsByName loads professors by name.
func (h *Handler) GetProfessorsByName(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Professor name is required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	professors, hasNext, err := h.db.GetProfessorsByName(c.Request.Context(), name, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get professors"})
		return
	}

	pagination := buildPaginationMeta(params, len(professors), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(professors),
		"professors": professors,
		"pagination": pagination,
	})
}
