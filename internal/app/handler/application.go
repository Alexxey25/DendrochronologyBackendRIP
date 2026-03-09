package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetDendrochronology(ctx *gin.Context) {
	draftApp, err := h.Repository.GetDraftApplication(creatorID)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	_, views, err := h.Repository.GetApplicationWithConstructions(draftApp.ID)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	totalSamples := h.Repository.GetTotalSamples(draftApp.ID)
	buildYear := h.Repository.GetEstimatedBuildYear(draftApp.ID)

	ctx.HTML(http.StatusOK, "dendrochronologypage.html", gin.H{
		"dendrochronology": draftApp,
		"constructions":    views,
		"totalSamples":     totalSamples,
		"buildYear":        buildYear,
		"minioBase":        minioBaseURL,
	})
}

func (h *Handler) AddToDendrochronology(ctx *gin.Context) {
	strId := ctx.PostForm("construction_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.Repository.AddConstructionToApplication(uint(id), creatorID)
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

func (h *Handler) UpdateSamplesCount(ctx *gin.Context) {
	strId := ctx.PostForm("item_id")
	itemID, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	delta := 1
	if ctx.PostForm("action") == "decrease" {
		delta = -1
	}

	err = h.Repository.UpdateSamplesCount(uint(itemID), delta)
	if err != nil {
		logrus.Error(err)
	}

	ctx.Redirect(http.StatusFound, "/dendrochronology")
}

func (h *Handler) FormApplication(ctx *gin.Context) {
	strId := ctx.PostForm("dendrochronology_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.Repository.FormApplication(uint(id))
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

func (h *Handler) DeleteDendrochronology(ctx *gin.Context) {
	strId := ctx.PostForm("dendrochronology_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.Repository.DeleteApplicationBySQL(uint(id))
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}
