package repository

import (
	"car-rental/internal/models"
	"gorm.io/gorm"
)

type CommissionRepositoryImpl struct {
	DB *gorm.DB
}

func NewCommissionRepositoryImpl(db *gorm.DB) *CommissionRepositoryImpl {
	return &CommissionRepositoryImpl{DB: db}
}

func (r CommissionRepositoryImpl) GetCommissionsByType(autoType string) []models.Commission {
	var commissions []models.Commission
	r.DB.Where("auto_type = ?", autoType).Find(&commissions)
	return commissions
}
