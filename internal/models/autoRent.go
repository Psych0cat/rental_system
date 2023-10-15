package models

import (
	"time"
)

type AutoRent struct {
	AutoID    string    `db:"auto_id"`
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
}

func (a *AutoRent) TableName() string {
	return "auto_rent"
}
