package handlers

import (
	"net/http"

	"github.com/acmutd/acmutd-api/internal/types"
	"github.com/gin-gonic/gin"
)

// GetGradesByProfID loads grade distributions by professor ID.
func (h *Handler) GetGradesByProfID(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Professor ID is required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	grades, hasNext, err := h.db.GetGradesByProfId(c.Request.Context(), id, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesByProfName loads grade distributions by professor name.
func (h *Handler) GetGradesByProfName(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Professor name is required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	grades, hasNext, err := h.db.GetGradesByProfName(c.Request.Context(), name, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesByTerm loads grade distributions by term.
func (h *Handler) GetGradesByTerm(c *gin.Context) {
	term := c.Param("term")

	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term is required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	grades, hasNext, err := h.db.GetGradesByTerm(c.Request.Context(), term, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesByPrefix loads grade distributions by prefix with optional term filter.
func (h *Handler) GetGradesByPrefix(c *gin.Context) {
	prefix := c.Param("prefix")

	if prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prefix is required"})
		return
	}

	term := c.Query("term")

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	var (
		grades  []types.Grades
		hasNext bool
		err     error
	)

	if term == "" {
		grades, hasNext, err = h.db.GetGradesByPrefix(c.Request.Context(), prefix, params.Limit, params.Offset)
	} else {
		grades, hasNext, err = h.db.GetGradesByPrefixAndTerm(c.Request.Context(), prefix, term, params.Limit, params.Offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesByPrefixAndNumber loads grade distributions by prefix and course number with optional term filter.
func (h *Handler) GetGradesByPrefixAndNumber(c *gin.Context) {
	prefix := c.Param("prefix")
	number := c.Param("number")

	if prefix == "" || number == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prefix and number are required"})
		return
	}

	term := c.Query("term")

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	var (
		grades  []types.Grades
		hasNext bool
		err     error
	)

	if term == "" {
		grades, hasNext, err = h.db.GetGradesByPrefixAndNumber(c.Request.Context(), prefix, number, params.Limit, params.Offset)
	} else {
		grades, hasNext, err = h.db.GetGradesByNumberAndTerm(c.Request.Context(), term, prefix, number, params.Limit, params.Offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesBySection loads grade distributions for specific section
func (h *Handler) GetGradesBySection(c *gin.Context) {
	term := c.Param("term")
	prefix := c.Param("prefix")
	number := c.Param("number")
	section := c.Param("section")

	if term == "" || prefix == "" || number == "" || section == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term, prefix, number, and section are required"})
		return
	}

	grades, err := h.db.GetGradesBySection(c.Request.Context(), term, prefix, number, section)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"grades": grades,
	})
}
