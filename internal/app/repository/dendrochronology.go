package repository

import (
	"time"

	"github.com/sirupsen/logrus"
	"metoda/internal/app/ds"
)

func (r *Repository) ClearDraftDendrochronologiesOnStartup() {
	r.db.Exec("UPDATE dendrochronologies SET status = ? WHERE status = ?", ds.StatusDeleted, ds.StatusDraft)
}

func (r *Repository) GetDraftDendrochronology(creatorID uint) (*ds.Dendrochronology, error) {
	var d ds.Dendrochronology
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) GetDendrochronologyWithConstructions(dendrochronologyID uint) (*ds.Dendrochronology, []DendrochronologyConstructionView, error) {
	var d ds.Dendrochronology
	err := r.db.First(&d, dendrochronologyID).Error
	if err != nil {
		return nil, nil, err
	}

	if d.Status == ds.StatusDeleted {
		return nil, nil, nil
	}

	var items []ds.DendrochronologyConstruction
	err = r.db.Where("dendrochronology_id = ?", dendrochronologyID).Preload("Construction").Order("id").Find(&items).Error
	if err != nil {
		return nil, nil, err
	}

	var views []DendrochronologyConstructionView
	for _, item := range items {
		views = append(views, DendrochronologyConstructionView{
			ID:                item.ID,
			ConstructionTitle: item.Construction.ConstructionTitle,
			UseLife:           item.Construction.UseLife,
			ImageURL:          item.Construction.ImageURL,
			SamplesCount:      item.SamplesCount,
			CuttingDate:       item.CuttingDate,
			DateCorrection:    item.DateCorrection,
		})
	}

	return &d, views, nil
}

type DendrochronologyConstructionView struct {
	ID                uint
	ConstructionTitle string
	UseLife           string
	ImageURL          string
	SamplesCount      int
	CuttingDate       string
	DateCorrection    string
}

// GetEstimatedBuildYear — предварительная дата постройки = max(cutting_date + date_correction)
// по строкам связи dendrochronology_constructions (по ER cutting_date и date_correction в этой таблице).
func (r *Repository) GetEstimatedBuildYear(dendrochronologyID uint) int {
	var year int

	r.db.Table("dendrochronology_constructions").
		Where("dendrochronology_id = ? AND cutting_date <> '' AND cutting_date IS NOT NULL AND date_correction <> '' AND date_correction IS NOT NULL", dendrochronologyID).
		Select("MAX((cutting_date)::int + (date_correction)::int)").
		Scan(&year)

	return year
}

func (r *Repository) AddConstructionToDendrochronology(constructionID uint, creatorID uint) error {
	var d ds.Dendrochronology
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&d).Error
	if err != nil {
		d = ds.Dendrochronology{
			Status:     ds.StatusDraft,
			DateCreate: time.Now(),
			CreatorID:  creatorID,
		}
		if err := r.db.Create(&d).Error; err != nil {
			return err
		}
	}

	dc := ds.DendrochronologyConstruction{
		DendrochronologyID: d.ID,
		ConstructionID:     constructionID,
		SamplesCount:       1,
	}
	return r.db.Create(&dc).Error
}

func (r *Repository) DeleteDendrochronologyBySQL(dendrochronologyID uint) error {
	result := r.db.Exec("UPDATE dendrochronologies SET status = 'удалён' WHERE id = ?", dendrochronologyID)
	return result.Error
}

func (r *Repository) UpdateSamplesCount(itemID uint, delta int) error {
	var item ds.DendrochronologyConstruction
	if err := r.db.First(&item, itemID).Error; err != nil {
		return err
	}

	newCount := item.SamplesCount + delta
	if newCount < 1 {
		newCount = 1
	}

	return r.db.Model(&item).Update("samples_count", newCount).Error
}

func (r *Repository) FormDendrochronology(dendrochronologyID uint) error {
	totalSamples := r.GetTotalSamples(dendrochronologyID)
	now := time.Now()

	return r.db.Model(&ds.Dendrochronology{}).Where("id = ?", dendrochronologyID).Updates(map[string]interface{}{
		"status":        ds.StatusFormed,
		"date_formed":   now,
		"total_samples": totalSamples,
	}).Error
}

func (r *Repository) GetTotalSamples(dendrochronologyID uint) int {
	var total int
	r.db.Model(&ds.DendrochronologyConstruction{}).
		Where("dendrochronology_id = ?", dendrochronologyID).
		Select("COALESCE(SUM(samples_count), 0)").
		Scan(&total)
	return total
}

func (r *Repository) GetCartCount(creatorID uint) int64 {
	var dendrochronologyID uint
	var count int64

	err := r.db.Model(&ds.Dendrochronology{}).Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).Select("id").First(&dendrochronologyID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.DendrochronologyConstruction{}).Where("dendrochronology_id = ?", dendrochronologyID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records:", err)
	}

	return count
}

func (r *Repository) GetDraftDendrochronologyID(creatorID uint) uint {
	var dendrochronologyID uint
	err := r.db.Model(&ds.Dendrochronology{}).Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).Select("id").First(&dendrochronologyID).Error
	if err != nil {
		return 0
	}
	return dendrochronologyID
}
