package repository

import (
	"time"

	"github.com/sirupsen/logrus"
	"metoda/internal/app/ds"
)

func (r *Repository) ClearDraftApplicationsOnStartup() {
	r.db.Exec("UPDATE applications SET status = ? WHERE status = ?", ds.StatusDeleted, ds.StatusDraft)
}

func (r *Repository) GetDraftApplication(creatorID uint) (*ds.Application, error) {
	var app ds.Application
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *Repository) GetApplicationWithConstructions(applicationID uint) (*ds.Application, []ApplicationConstructionView, error) {
	var app ds.Application
	err := r.db.First(&app, applicationID).Error
	if err != nil {
		return nil, nil, err
	}

	if app.Status == ds.StatusDeleted {
		return nil, nil, nil
	}

	var items []ds.ApplicationConstruction
	err = r.db.Where("application_id = ?", applicationID).Preload("Construction").Order("id").Find(&items).Error
	if err != nil {
		return nil, nil, err
	}

	var views []ApplicationConstructionView
	for _, item := range items {
		views = append(views, ApplicationConstructionView{
			ID:                item.ID,
			ConstructionTitle: item.Construction.ConstructionTitle,
			UseLife:           item.Construction.UseLife,
			ImageURL:          item.Construction.ImageURL,
			SamplesCount:      item.SamplesCount,
			CuttingDate:       item.Construction.CuttingDate,
			DateCorrection:    item.Construction.DateCorrection,
		})
	}

	return &app, views, nil
}

type ApplicationConstructionView struct {
	ID                uint
	ConstructionTitle string
	UseLife           string
	ImageURL          string
	SamplesCount      int
	CuttingDate       string
	DateCorrection    string
}

// GetEstimatedBuildYear рассчитывает предварительную дату постройки заявки
// как max(cutting_date + date_correction) по всем связанным конструкциям.
// cutting_date и date_correction хранятся как строки с годом и поправкой в годах.
func (r *Repository) GetEstimatedBuildYear(applicationID uint) int {
	var year int

	r.db.Table("application_constructions AS ac").
		Joins("JOIN constructions c ON c.id = ac.construction_id").
		Where("ac.application_id = ? AND c.cutting_date <> '' AND c.date_correction <> ''", applicationID).
		Select("MAX((c.cutting_date)::int + (c.date_correction)::int)").
		Scan(&year)

	return year
}

func (r *Repository) AddConstructionToApplication(constructionID uint, creatorID uint) error {
	var app ds.Application
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&app).Error
	if err != nil {
		app = ds.Application{
			Status:     ds.StatusDraft,
			DateCreate: time.Now(),
			CreatorID:  creatorID,
		}
		if err := r.db.Create(&app).Error; err != nil {
			return err
		}
	}

	ac := ds.ApplicationConstruction{
		ApplicationID:  app.ID,
		ConstructionID: constructionID,
		SamplesCount:   1,
	}
	return r.db.Create(&ac).Error
}

func (r *Repository) DeleteApplicationBySQL(applicationID uint) error {
	result := r.db.Exec("UPDATE applications SET status = 'удалён' WHERE id = ?", applicationID)
	return result.Error
}

func (r *Repository) UpdateSamplesCount(itemID uint, delta int) error {
	var item ds.ApplicationConstruction
	if err := r.db.First(&item, itemID).Error; err != nil {
		return err
	}

	newCount := item.SamplesCount + delta
	if newCount < 1 {
		newCount = 1
	}

	return r.db.Model(&item).Update("samples_count", newCount).Error
}

func (r *Repository) FormApplication(applicationID uint) error {
	totalSamples := r.GetTotalSamples(applicationID)
	now := time.Now()

	return r.db.Model(&ds.Application{}).Where("id = ?", applicationID).Updates(map[string]interface{}{
		"status":        ds.StatusFormed,
		"date_formed":   now,
		"total_samples": totalSamples,
	}).Error
}

func (r *Repository) GetTotalSamples(applicationID uint) int {
	var total int
	r.db.Model(&ds.ApplicationConstruction{}).
		Where("application_id = ?", applicationID).
		Select("COALESCE(SUM(samples_count), 0)").
		Scan(&total)
	return total
}

func (r *Repository) GetCartCount(creatorID uint) int64 {
	var applicationID uint
	var count int64

	err := r.db.Model(&ds.Application{}).Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).Select("id").First(&applicationID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.ApplicationConstruction{}).Where("application_id = ?", applicationID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records:", err)
	}

	return count
}

func (r *Repository) GetDraftApplicationID(creatorID uint) uint {
	var applicationID uint
	err := r.db.Model(&ds.Application{}).Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).Select("id").First(&applicationID).Error
	if err != nil {
		return 0
	}
	return applicationID
}
