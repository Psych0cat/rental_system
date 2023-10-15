package repository

import "car-rental/internal/models"

type AutoRepository interface {
	GetAvailableAutoByType(autoType string) ([]models.Auto, error)
	GetAutoById(autoId string) (models.Auto, error)
	BindAuto(autoId string) error
	ReleaseAuto(autoId string) error
}
