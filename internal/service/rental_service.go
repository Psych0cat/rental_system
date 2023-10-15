package service

import (
	"time"

	"car-rental/internal/models"
)

type RentalService interface {
	GetAvailableAutoByType(autoType string) ([]models.Auto, error)
	BindAuto(autoId string, days int) error
	ReleaseAuto(autoId string, releaseDate time.Time) (checkout int, err error)
	GetCurrentCommission(autoId string, calculationDate time.Time) (commission int, insurance int, err error)
}
