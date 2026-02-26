package repository

import (
	"metoda/internal/app/ds"
)

func (r *Repository) GetAllConstructions() ([]ds.Construction, error) {
	var constructions []ds.Construction
	err := r.db.Where("is_delete = false").Find(&constructions).Error
	if err != nil {
		return nil, err
	}
	return constructions, nil
}

func (r *Repository) GetConstructionByID(id int) (*ds.Construction, error) {
	var construction ds.Construction
	err := r.db.Where("id = ? AND is_delete = false", id).First(&construction).Error
	if err != nil {
		return nil, err
	}
	return &construction, nil
}

func (r *Repository) SearchConstructionsByTitle(title string) ([]ds.Construction, error) {
	var constructions []ds.Construction
	err := r.db.Where("construction_title ILIKE ? AND is_delete = ?", "%"+title+"%", false).Find(&constructions).Error
	if err != nil {
		return nil, err
	}
	return constructions, nil
}

func (r *Repository) DeleteConstruction(constructionID uint) error {
	return r.db.Model(&ds.Construction{}).Where("id = ?", constructionID).UpdateColumn("is_delete", true).Error
}
