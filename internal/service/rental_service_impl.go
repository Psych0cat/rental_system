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
	total := 0
	weekendComm := 0
	dailyCommission, weekendCommission,
		insuranceCommission, agreementCommission, penaltyPercentCommission := getCommissions(commissions)
	if checkout && penaltyPercentCommission.Value != 0 && penaltyPercentCommission.MinThreshold != 0 {
		return calculateCheckouts(
			rent, releaseDate, dailyCommission, weekendCommission, agreementCommission, insuranceCommission, penaltyPercentCommission)
	} else {
		// get current commission, same as checkout without commission
		newD1 := rent.StartDate.Truncate(time.Hour * 24)
		newD2 := releaseDate.Truncate(time.Hour * 24)
		complete := int(math.Ceil(newD2.Sub(newD1).Hours()/24)) + 1
		// first day always counts as full day
		if !checkout && complete == 0 {
			complete = 1
		}
		_, weekEnd := calculateWeekends(rent.StartDate, complete)
		if dailyCommission != 0 {
			daily := calculateDailyCommission(complete, dailyCommission)
			total += daily
		}
		if weekendCommission != 0 {
			weekendComm = calculateWeekendCommission(weekEnd, dailyCommission, weekendCommission)
			total += weekendComm
		}
		return total + agreementCommission, insuranceCommission
	}
}

func getCommissions(commissions []models.Commission) (
	dailyCommission,
	weekendCommission,
	insuranceCommission,
	agreementCommission int,
	penaltyPercentCommission models.Commission) {
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
	return dailyCommission, weekendCommission, insuranceCommission, agreementCommission, penaltyPercentCommission
}

func calculateCheckouts(
	rent models.AutoRent,
	releaseDate time.Time,
	dailyCommission,
	weekendCommission,
	agreementCommission,
	insuranceCommission int,
	penaltyPercentCommission models.Commission) (int, int) {
	daily := 0
	total := 0
	rent.EndDate = rent.EndDate.AddDate(0, 0, 1)
	if penaltyPercentCommission.MinThreshold != 0 {
		complete, left := calculateDays(
			rent.StartDate, rent.EndDate,
			releaseDate, penaltyPercentCommission.MinThreshold)
		_, weekEnd := calculateWeekends(rent.StartDate, complete)
		if dailyCommission != 0 {
			daily = calculateDailyCommission(complete, dailyCommission)
			total += daily
		}
		if weekendCommission != 0 {
			weekendComm := calculateWeekendCommission(weekEnd, dailyCommission, weekendCommission)
			total += weekendComm
		}
		if left == 0 {
			return total + agreementCommission, insuranceCommission
		} else {
			t := rent.StartDate.AddDate(0, 0, penaltyPercentCommission.MinThreshold-1)
			t = t.AddDate(0, 0, 2)
			penaltyCostBeforeCommission := 0
			_, weekEnd = calculateWeekends(t, left)
			penaltyCostBeforeCommission += calculateDailyCommission(left, dailyCommission)
			penaltyCostBeforeCommission += calculateWeekendCommission(weekEnd, dailyCommission, weekendCommission)
			finPenalty := calculatePenaltyCommission(
				penaltyCostBeforeCommission, penaltyPercentCommission.Value)
			return total + finPenalty + agreementCommission, insuranceCommission
		}
	}
	return 0, 0
}

func calculateDays(startDate time.Time, endDate time.Time, releaseDate time.Time, threshold int) (completeDays int, daysLeft int) {
	completeDays = 0
	left := 0
	newD1 := startDate.Truncate(time.Hour * 24)
	newD2 := releaseDate.Truncate(time.Hour * 24)
	completeDays = int(math.Ceil(newD2.Sub(newD1).Hours()/24)) + 1
	if completeDays < threshold {
		completeDays = threshold
		releaseDate = startDate.AddDate(0, 0, threshold)
		newD1 = releaseDate.Truncate(time.Hour * 24)
		newD2 = endDate.Truncate(time.Hour * 24)
		daysLeft = int(endDate.Sub(releaseDate).Hours() / 24)
		left = int(endDate.Sub(startDate.AddDate(0, 0, completeDays)).Hours() / 24)
		return completeDays, left - 1
	}
	newD1 = releaseDate.Truncate(time.Hour * 24)
	newD2 = endDate.Truncate(time.Hour * 24)
	daysLeft = int(newD2.Sub(newD1).Hours() / 24)
	return completeDays, daysLeft - 1
}

func calculateWeekends(startDate time.Time, numDays int) (workDay, weekendDays int) {
	workDay = 0
	weekendDays = 0
	for i := 0; i < numDays; i++ {
		if startDate.Weekday() == time.Saturday || startDate.Weekday() == time.Sunday {
			weekendDays++
		} else {
			workDay++
		}
		startDate = startDate.AddDate(0, 0, 1)
	}
	return workDay, weekendDays
}

func calculateDailyCommission(completeDays int, dailyCommission int) int {
	return completeDays * dailyCommission
}

func calculateWeekendCommission(completeDays int, dailyCommission int, weekendCommission int) int {
	return completeDays * dailyCommission * weekendCommission / 100
}

func calculatePenaltyCommission(total int, percent int) int {
	return total * percent / 100
}
