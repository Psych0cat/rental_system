package repository

import "car-rental/internal/models"

type CommissionRepository interface {
	GetCommissionsByType(autoType string) []models.Commission
}
