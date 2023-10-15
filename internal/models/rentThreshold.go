package models

type RentThreshold struct {
	AutoType     string `db:"auto_type"`
	MinThreshold int    `db:"min_threshold"`
	MaxThreshold int    `db:"max_threshold"`
}

func (a *RentThreshold) TableName() string {
	return "rent_threshold"
}
