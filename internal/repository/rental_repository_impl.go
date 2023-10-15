package repository

import (
	"car-rental/internal/models"
	"errors"
	"gorm.io/gorm"
	"time"
)

type RentalRepositoryImpl struct {
	DB *gorm.DB
}

func NewRentalRepositoryImpl(db *gorm.DB) *RentalRepositoryImpl {
	return &RentalRepositoryImpl{DB: db}
}

func (r RentalRepositoryImpl) GetRentByAuto(autoId string) (models.AutoRent, error) {
	var rent models.AutoRent
	res := r.DB.Where("auto_id = ?", autoId).First(&rent)
	if res.Error != nil {
		return rent, res.Error
	}

	return rent, nil
}

func (r RentalRepositoryImpl) BindRent(autoId string, days int) error {
	var rent models.AutoRent
	rent.AutoID = autoId
	rent.StartDate = time.Now()
	rent.EndDate = time.Now().AddDate(0, 0, days)
	res := r.DB.Create(&rent)
	if res.Error != nil {
		return errors.New("auto not found")
	}
	return nil

}

func (r RentalRepositoryImpl) ReleaseRent(autoId string) error {
	var rent models.AutoRent
	res := r.DB.Where("auto_id = ?", autoId).Delete(&rent)
	return res.Error
}

func (r RentalRepositoryImpl) GetThresholdsByAutoType(autoType string) (models.RentThreshold, error) {
	var thresholds models.RentThreshold
	res := r.DB.Where("auto_type = ?", autoType).Find(&thresholds)
	if res.Error != nil {
		return thresholds, res.Error
	}
	return thresholds, nil
}
