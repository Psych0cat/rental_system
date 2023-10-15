package repository

import "car-rental/internal/models"

type RentalRepository interface {
	GetRentByAuto(autoId string) (models.AutoRent, error)
	BindRent(autoId string, days int) error
	ReleaseRent(autoId string) error
	GetThresholdsByAutoType(autoType string) (models.RentThreshold, error)
}
