package service

import (
	"errors"
	"fmt"
	"math"
	"time"

	"car-rental/internal/models"
	"car-rental/internal/repository"
)

const (
	commissionTypeDaily      = "daily"
	commissionTypeAgreement  = "agreement"
	commissionTypeWeekend    = "weekend"
	commissionTypePenalty    = "penalty"
	commissionTypeInsurance  = "insurance"
	NotFoundError            = "record not found"
	AlreadyRentedError       = "auto is already rented"
	ThresholdValidationError = "days should be between min and max threshold"
)

type RentalServiceImpl struct {
	autoRepository       repository.AutoRepository
	rentalRepository     repository.RentalRepository
	commissionRepository repository.CommissionRepository
}

func NewRentalServiceImpl(autoRepository repository.AutoRepository,
	rentalRepository repository.RentalRepository,
	commissionRepository repository.CommissionRepository) *RentalServiceImpl {
	return &RentalServiceImpl{
		autoRepository:       autoRepository,
		rentalRepository:     rentalRepository,
		commissionRepository: commissionRepository,
	}
}

func (a RentalServiceImpl) GetAvailableAutoByType(autoType string) ([]models.Auto, error) {
	autos, err := a.autoRepository.GetAvailableAutoByType(autoType)
	if err != nil {
		return nil, err
	}
	if len(autos) == 0 {
		return nil, errors.New(NotFoundError)
	}
	return autos, nil
}

func (a RentalServiceImpl) BindAuto(autoId string, days int) error {
	rent, err := a.rentalRepository.GetRentByAuto(autoId)
	if err != nil && err.Error() != NotFoundError {
		return err
	}
	if rent != (models.AutoRent{}) {
		return errors.New("auto is already rented")
	}
	auto, err := a.autoRepository.GetAutoById(autoId)
	if err != nil {
		return err
	}
	threshold, err := a.rentalRepository.GetThresholdsByAutoType(auto.Type)
	if err != nil {
		return err
	}
	if threshold != (models.RentThreshold{}) {
		if days < threshold.MinThreshold || days > threshold.MaxThreshold {
			return errors.New(fmt.Sprint(
				"days should be between ", threshold.MinThreshold, " and ", threshold.MaxThreshold))
		}
	}
	err = a.rentalRepository.BindRent(autoId, days)
	if err != nil {
		return err
	}
	err = a.autoRepository.BindAuto(autoId)
	if err != nil {
		return err
	}
	return nil
}

func (a RentalServiceImpl) ReleaseAuto(
	autoId string, releaseDate time.Time) (checkout int, err error) {
	rent, err := a.rentalRepository.GetRentByAuto(autoId)
	if err != nil {
		return 0, err
	}
	if rent == (models.AutoRent{}) {
		return 0, errors.New(NotFoundError)
	}
	auto, err := a.autoRepository.GetAutoById(autoId)
	commissions := a.commissionRepository.GetCommissionsByType(auto.Type)
	checkout, _ = calculateCommissions(rent, commissions, releaseDate, true)
	err = a.autoRepository.ReleaseAuto(autoId)
	if err != nil {
		return 0, err
	}
	err = a.rentalRepository.ReleaseRent(autoId)
	if err != nil {
		return 0, err
	}
	return checkout, nil
}

func (a RentalServiceImpl) GetCurrentCommission(autoId string, calculationDate time.Time) (
	commission int, insurance int, err error) {
	rent, err := a.rentalRepository.GetRentByAuto(autoId)
	if err != nil {
		return 0, 0, err
	}
	if rent == (models.AutoRent{}) {
		return 0, 0, err
	}
	auto, err := a.autoRepository.GetAutoById(autoId)
	if err != nil {
		return 0, 0, err
	}
	necessaryCommissions := a.commissionRepository.GetCommissionsByType(auto.Type)
	if necessaryCommissions == nil {
		return 0, 0, errors.New("no commissions found for auto type")
	}
	currentCommission, insurance := calculateCommissions(
		rent, necessaryCommissions, calculationDate.AddDate(0, 0, -1), false)
	return currentCommission, insurance, nil
}

// Left some flexibility, for example we can add weekend/penalty commission for standard auto
// or add commissions to new auto types via DB, without changing code
// left cases like penalty + businessday commissions without weekend commission out of scope to keep it short
func calculateCommissions(
	rent models.AutoRent, commissions []models.Commission, releaseDate time.Time, checkout bool) (
	totalCommission int, insuranceCommission int) {
	releaseDate = releaseDate.Round(0)
	newD1 := rent.StartDate.Truncate(time.Hour * 24)
	newD2 := releaseDate.Truncate(time.Hour * 24)
	completeDays := int(math.Ceil(newD2.Sub(newD1).Hours() / 24))
	if completeDays < 0 {
		completeDays = -completeDays
	}
	totalCost := 0
	dailyCommission := 0
	agreementCommission := 0
	weekendCommission := 0
	penaltyPercentCommission := models.Commission{}
	insuranceCommission = 0
	for _, commission := range commissions {
		if commission.Type == commissionTypeDaily {
			dailyCommission = commission.Value
		}
		if commission.Type == commissionTypeAgreement {
			agreementCommission = commission.Value
		}
		if commission.Type == commissionTypeWeekend {
			weekendCommission = commission.Value
		}
		if commission.Type == commissionTypePenalty {
			penaltyPercentCommission = commission
		}
		if commission.Type == commissionTypeInsurance {
			insuranceCommission = commission.Value
		}
	}

	if penaltyPercentCommission.Value != 0 && weekendCommission != 0 && dailyCommission != 0 && checkout {
		totalCost += calculatePenaltyCommissionWithWeekends(
			releaseDate, rent.StartDate, rent.EndDate, completeDays,
			penaltyPercentCommission, dailyCommission, weekendCommission)
		return totalCost + agreementCommission, insuranceCommission
	}
	// Currently used only for current balance calculations
	if weekendCommission != 0 && dailyCommission != 0 {
		if completeDays == 1 {
			oneDayCommission := handleOneDayCase(
				releaseDate.AddDate(0, 0, +1),
				dailyCommission, weekendCommission, agreementCommission)
			return oneDayCommission, insuranceCommission
		}
		totalDaily := calculateWithWeekendCommission(
			rent.StartDate.AddDate(0, 0, -1), completeDays+1,
			weekendCommission, dailyCommission)
		totalCost += totalDaily
		return totalCost + agreementCommission, insuranceCommission
	}
	if dailyCommission != 0 {
		totalCost += completeDays + 1*dailyCommission
		return totalCost + agreementCommission, insuranceCommission
	}

	return totalCost, insuranceCommission
}

func handleOneDayCase(checkDate time.Time,
	dailyCommission int, weekendCommission int, agreement int) int {
	if checkDate.Weekday() == time.Saturday || checkDate.Weekday() == time.Sunday {
		return 1*dailyCommission + 1*dailyCommission*weekendCommission/100 + agreement
	} else {
		return 1*dailyCommission + agreement
	}
}

func calculatePenaltyCommissionWithWeekends(
	releaseDate time.Time,
	startDate time.Time,
	endDate time.Time,
	completeDays int,
	penaltyCommission models.Commission,
	dailyCommission int,
	weekendCommission int) int {
	checkout := 0
	minDays := penaltyCommission.MinThreshold
	daysLeft := int(endDate.Sub(releaseDate).Hours()/24) + 1
	// If min days threshold is not reached
	if completeDays+1 < penaltyCommission.MinThreshold {
		left := int(endDate.Sub(startDate.AddDate(0, 0, minDays)).Hours() / 24)
		checkout += calculateWithWeekendCommission(
			startDate, minDays, weekendCommission, dailyCommission)
		penaltyForUnusedDays := calculateWithWeekendCommission(
			startDate.AddDate(0, 0, minDays+1), left, weekendCommission, dailyCommission)
		checkout += (penaltyForUnusedDays * penaltyCommission.Value) / 100
	} else if dailyCommission != 0 {
		checkout += calculateWithWeekendCommission(
			startDate, completeDays+1, weekendCommission, dailyCommission)
		unusedDays := calculateWithWeekendCommission(
			releaseDate.AddDate(0, 0, 1),
			daysLeft, weekendCommission, dailyCommission)
		penalty := (unusedDays * penaltyCommission.Value) / 100
		checkout += penalty
	}
	return checkout
}

// calculate businessday and weekend cost separately
func calculateWithWeekendCommission(startDate time.Time, completeDays int,
	weekendCommission int, dailyCommission int) int {
	if completeDays == 0 {
		return 0
	}
	businessDaysComplete := getWeekdaysBetween(
		startDate, startDate.AddDate(0, 0, completeDays))
	weekendDaysComplete := completeDays - businessDaysComplete
	standardCost := weekendDaysComplete * dailyCommission
	actualWeekendCommission := (standardCost * weekendCommission) / 100
	weekendCost := standardCost + actualWeekendCommission
	businessDayCostDaily := businessDaysComplete * dailyCommission
	return businessDayCostDaily + weekendCost
}

func getWeekdaysBetween(t, f time.Time) int {
	days := 0
	for {
		if t.Equal(f) {
			return days
		}
		if t.Weekday() != time.Saturday && t.Weekday() != time.Sunday {
			days++
		}
		t = t.Add(time.Hour * 24)
	}
}
