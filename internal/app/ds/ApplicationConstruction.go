package ds

type ApplicationConstruction struct {
	ID             uint `gorm:"primaryKey"`
	ApplicationID  uint `gorm:"not null;uniqueIndex:idx_app_construction"`
	ConstructionID uint `gorm:"not null;uniqueIndex:idx_app_construction"`
	SamplesCount   int  `gorm:"not null;default:1"`

	Application  Application  `gorm:"foreignKey:ApplicationID"`
	Construction Construction `gorm:"foreignKey:ConstructionID"`
}
