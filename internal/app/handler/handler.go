package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"metoda/internal/app/repository"
)

const minioBaseURL = "http://localhost:9000/constructions"

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetConstructions(ctx *gin.Context) {
	var constructions []repository.Construction
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		constructions, err = h.Repository.GetConstructions()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		constructions, err = h.Repository.GetConstructionsByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	applications, _ := h.Repository.GetApplications()

	ctx.HTML(http.StatusOK, "mainpage.html", gin.H{
		"constructions": constructions,
		"query":         searchQuery,
		"cartCount":     len(applications),
		"minioBase":     minioBaseURL,
	})
}

func (h *Handler) GetConstruction(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	construction, err := h.Repository.GetConstruction(id)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	ctx.HTML(http.StatusOK, "productpage.html", gin.H{
		"construction": construction,
		"minioBase":    minioBaseURL,
	})
}

func (h *Handler) GetApplications(ctx *gin.Context) {
	applications, err := h.Repository.GetApplications()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "applicationpage.html", gin.H{
		"applications": applications,
		"minioBase":    minioBaseURL,
	})
}
