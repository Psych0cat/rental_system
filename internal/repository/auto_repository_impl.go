package repository

import (
	"car-rental/internal/models"
	"gorm.io/gorm"
)

type AutoRepositoryImpl struct {
	DB *gorm.DB
}

func NewAutoRepositoryImpl(db *gorm.DB) *AutoRepositoryImpl {
	return &AutoRepositoryImpl{DB: db}
}

func (a AutoRepositoryImpl) GetAvailableAutoByType(autoType string) ([]models.Auto, error) {
	var auto []models.Auto
	filter := &models.Auto{
		Type:         autoType,
		Availability: true,
	}
	res := a.DB.Where(filter).Find(&auto)
	if res.Error != nil {
		return nil, res.Error
	}
	return auto, nil
}

func (a AutoRepositoryImpl) GetAutoById(autoId string) (models.Auto, error) {
	var auto models.Auto
	res := a.DB.Where("id = ?", autoId).First(&auto)
	if res.Error != nil {
		return auto, res.Error
	}
	return auto, nil
}

func (a AutoRepositoryImpl) BindAuto(autoId string) error {
	var auto models.Auto
	res := a.DB.Where("id = ?", autoId).First(&auto)
	if res.Error != nil {
		return res.Error
	}
	auto.Availability = false
	a.DB.Save(&auto)
	return nil
}

func (a AutoRepositoryImpl) ReleaseAuto(autoId string) error {
	var auto models.Auto
	res := a.DB.Where("id = ?", autoId).First(&auto)
	if res.Error != nil {
		return res.Error
	}
	auto.Availability = true
	a.DB.Save(&auto)
	return nil
}
