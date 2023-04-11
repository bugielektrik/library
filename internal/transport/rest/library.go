package rest

import (
	"library/internal/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateAuthor(c *gin.Context) {
	req := dto.AuthorRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusUnsupportedMediaType, err.Error())
		return
	}

	err := h.validate.Struct(req)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.libraries.CreateAuthor(req)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *Handler) GetAuthor(c *gin.Context) {
	id := c.Param("id")

	res, err := h.libraries.GetAuthor(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}
func (h *Handler) GetAuthors(c *gin.Context) {
	res, err := h.libraries.GetAuthors()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) UpdateAuthor(c *gin.Context) {
	req := dto.AuthorRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusUnsupportedMediaType, err.Error())
		return
	}
	req.ID = c.Param("id")

	err := h.validate.Struct(req)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err = h.libraries.UpdateAuthor(req)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, req)
}

func (h *Handler) DeleteAuthor(c *gin.Context) {
	id := c.Param("id")

	err := h.libraries.DeleteAuthor(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateBook(c *gin.Context) {
	req := dto.BookRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusUnsupportedMediaType, err.Error())
		return
	}

	err := h.validate.Struct(req)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.libraries.CreateBook(req)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *Handler) GetBook(c *gin.Context) {
	id := c.Param("id")

	res, err := h.libraries.GetBook(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) GetBooks(c *gin.Context) {
	res, err := h.libraries.GetBooks()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) UpdateBook(c *gin.Context) {
	req := dto.BookRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusUnsupportedMediaType, err.Error())
		return
	}
	req.ID = c.Param("id")

	err := h.validate.Struct(req)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err = h.libraries.UpdateBook(req)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, req)
}

func (h *Handler) DeleteBook(c *gin.Context) {
	id := c.Param("id")

	err := h.libraries.DeleteBook(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateMember(c *gin.Context) {
	req := dto.MemberRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusUnsupportedMediaType, err.Error())
		return
	}

	err := h.validate.Struct(req)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.libraries.CreateMember(req)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *Handler) GetMember(c *gin.Context) {
	id := c.Param("id")

	data, err := h.libraries.GetMember(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetMembers(c *gin.Context) {
	data, err := h.libraries.GetMembers()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *Handler) UpdateMember(c *gin.Context) {
	req := dto.MemberRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusUnsupportedMediaType, err.Error())
		return
	}
	req.ID = c.Param("id")

	err := h.validate.Struct(req)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err = h.libraries.UpdateMember(req)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, req)
}

func (h *Handler) DeleteMember(c *gin.Context) {
	id := c.Param("id")

	err := h.libraries.DeleteMember(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
