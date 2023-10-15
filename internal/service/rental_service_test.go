package service

import (
	"car-rental/internal/models"
	"car-rental/internal/repository"
	"car-rental/internal/utils"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestGetAvailableAutoByType(t *testing.T) {
	db, svc, err := setupRentServiceTests()
	if err != nil {
		t.Error(err)
	}
	db.Create(&models.AutoType{
		ID: "TestGetAvailableAutoByType",
	},
	)
	db.Create(&models.AutoType{
		ID: "TestGetAvailableAutoByType2",
	},
	)
	db.Create(&models.Auto{
		ID:           "auto1",
		Type:         "TestGetAvailableAutoByType",
		Availability: true,
	})
	db.Create(&models.Auto{
		ID:           "auto3",
		Type:         "TestGetAvailableAutoByType",
		Availability: true,
	})

	auto, err := svc.GetAvailableAutoByType("TestGetAvailableAutoByType")
	if err != nil {
		t.Error(err)
	}
	if len(auto) != 2 {
		t.Error("expected 2 autos, got ", len(auto))
	}
	db.Delete(&models.Auto{}, "type = ?", "TestGetAvailableAutoByType")
	db.Delete(&models.Auto{}, "type = ?", "TestGetAvailableAutoByType2")
	db.Delete(&models.AutoType{}, "id = ?", "TestGetAvailableAutoByType")
	db.Delete(&models.AutoType{}, "id = ?", "TestGetAvailableAutoByType2")
}

func TestBindAuto(t *testing.T) {
	db, svc, err := setupRentServiceTests()
	if err != nil {
		t.Error(err)
	}
	db.Create(&models.AutoType{
		ID: "TestBindAuto",
	},
	)
	db.Create(&models.Auto{
		ID:   "TestBindAuto",
		Type: "TestBindAuto",
	})
	err = svc.BindAuto("TestBindAuto", 10)
	if err != nil {
		t.Error(err)
	}
	_, err = svc.ReleaseAuto("TestBindAuto", time.Now())
	if err != nil {
		t.Error(err)
	}
	db.Delete(&models.Auto{}, "type = ?", "TestBindAuto")
	db.Delete(&models.AutoType{}, "id = ?", "TestBindAuto")
}

func TestReleaseAuto(t *testing.T) {
	db, svc, err := setupRentServiceTests()
	if err != nil {
		t.Error(err)
	}
	// release unexisting auto
	_, err = svc.ReleaseAuto("TESTAUTO", time.Now())
	if err.Error() != "record not found" {
		t.Error("expected error, got nil")
	}
	db.Create(&models.AutoType{
		ID: "test",
	},
	)

	db.Create(&models.Commission{
		AutoType: "test",
		Type:     commissionTypeDaily,
		Value:    200,
	})
	db.Create(&models.Commission{
		AutoType: "test",
		Type:     commissionTypeWeekend,
		Value:    20,
	})
	db.Create(&models.Commission{
		AutoType: "test",
		Type:     commissionTypeAgreement,
		Value:    200,
	})
	db.Create(&models.Commission{
		AutoType:     "test",
		Type:         commissionTypePenalty,
		Value:        5,
		MinThreshold: 10,
	})

	{ // min rent threshold  met, 10 full days + 3 weekdays penalty
		testday := time.Date(2023, time.October, 1, 1, 2, 3, 4, time.UTC)
		db.Create(&models.Auto{
			ID:           "TESTAUTO1",
			Type:         "test",
			Availability: false,
		})
		db.Create(models.AutoRent{
			AutoID:    "TESTAUTO1",
			StartDate: testday.AddDate(0, 0, -9),
			EndDate:   testday.AddDate(0, 0, 3),
		})
		// 6 working + 4 we + penalty for 3 working + agreement
		want := (6*200 + 4*200 + 4*200*20/100) + (3*200*5)/100 + 200
		commission, err1 := svc.ReleaseAuto("TESTAUTO1", testday)
		if err1 != nil {
			t.Error(err1)
		}
		if commission != want {
			t.Errorf("want %d, got %d", want, commission)
		}
		db.Delete(&models.Auto{}, "type = ?", "test")
	}
	{
		// min rent threshold met, wd and we left
	}
	testday := time.Date(2023, time.November, 10, 1, 2, 3, 4, time.UTC)
	db.Create(&models.Auto{
		ID:           "TESTAUTO2",
		Type:         "test",
		Availability: false,
	})
	db.Create(models.AutoRent{
		AutoID:    "TESTAUTO2",
		StartDate: testday.AddDate(0, 0, -12),
		EndDate:   testday.AddDate(0, 0, 10),
	})
	// 10 working + 3 we + weekend percent (6wd + 4we) + agreement
	want := (13 * 200) + 3*200*20/100 + ((10*200 + 4*200*20/100) * 5 / 100) + 200
	commission, err1 := svc.ReleaseAuto("TESTAUTO2", testday)
	if err1 != nil {
		t.Error(err1)
	}
	if commission != want {
		t.Errorf("want %d, got %d", want, commission)
	}

	{ // min rent threshold not met, 10 full + 2 weekdays penalty
		testday = time.Date(2023, time.November, 15, 1, 2, 3, 4, time.UTC)
		db.Create(&models.Auto{
			ID:           "TESTAUTO3",
			Type:         "test",
			Availability: false,
		})
		db.Create(models.AutoRent{
			AutoID:    "TESTAUTO3",
			StartDate: testday.AddDate(0, 0, -5),
			EndDate:   testday.AddDate(0, 0, 7),
		})
		//
		want := (10 * 200) + (4 * 200 * 20 / 100) + (2*200*5)/100 + 200
		commission, err1 = svc.ReleaseAuto("TESTAUTO3", testday)
		if err1 != nil {
			t.Error(err1)
		}
		if commission != want {
			t.Errorf("want %d, got %d", want, commission)
		}
		{ // rented and canceled in one day
			testday = time.Now()
			db.Create(&models.Auto{
				ID:           "TESTAUTO4",
				Type:         "test",
				Availability: false,
			})
			db.Create(models.AutoRent{
				AutoID:    "TESTAUTO4",
				StartDate: testday,
				EndDate:   testday.AddDate(0, 0, 10),
			})
			//
			want := 10*200 + (3 * 200 * 20 / 100) + 200
			commission, err1 = svc.ReleaseAuto("TESTAUTO4", testday)
			if err1 != nil {
				t.Error(err1)
			}
			if commission != want {
				t.Errorf("want %d, got %d", want, commission)
			}
		}

	}
	db.Delete(&models.Commission{}, "auto_type = ?", "test")
	db.Delete(&models.Auto{}, "type = ?", "test")
	db.Delete(&models.AutoType{}, "id = ?", "test")
}

func setupRentServiceTests() (*gorm.DB, *RentalServiceImpl, error) {
	dsn := utils.GetDsnFromEnv()
	db, err := models.ConnectDatabase(dsn)
	if err != nil {
		return nil, nil, err
	}
	ar := repository.NewAutoRepositoryImpl(db)
	cr := repository.NewCommissionRepositoryImpl(db)
	rr := repository.NewRentalRepositoryImpl(db)
	svc := NewRentalServiceImpl(ar, rr, cr)
	return db, svc, nil
}

func TestGetCurrentCommission(t *testing.T) {
	db, svc, err := setupRentServiceTests()
	if err != nil {
		t.Error(err)
	}
	db.Create(&models.AutoType{
		ID: "TestGetCurrentCommission",
	})
	db.Create(&models.Auto{ID: "TestGetCurrentCommission", Type: "TestGetCurrentCommission"})
	db.Create(&models.Commission{
		AutoType: "TestGetCurrentCommission",
		Type:     commissionTypeDaily,
		Value:    200,
	})
	db.Create(&models.Commission{
		AutoType: "TestGetCurrentCommission",
		Type:     commissionTypeWeekend,
		Value:    20,
	})
	db.Create(&models.Commission{
		AutoType: "TestGetCurrentCommission",
		Type:     commissionTypeAgreement,
		Value:    200,
	})
	db.Create(&models.Commission{
		AutoType: "TestGetCurrentCommission",
		Type:     "insurance",
		Value:    200,
	})
	{
	}
	{
		db.Create(&models.Commission{
			AutoType: "TestGetCurrentCommission1",
			Type:     commissionTypeDaily,
			Value:    200,
		})
		db.Create(&models.Commission{
			AutoType: "TestGetCurrentCommission1",
			Type:     commissionTypeWeekend,
			Value:    20,
		})
		db.Create(&models.Commission{
			AutoType: "TestGetCurrentCommission1",
			Type:     commissionTypeAgreement,
			Value:    200,
		})
		db.Create(&models.Commission{
			AutoType:     "TestGetCurrentCommission1",
			Type:         "commissionTypePenalty",
			Value:        5,
			MinThreshold: 10,
		})
		testday := time.Date(2023, time.November, 15, 1, 2, 3, 4, time.UTC)
		db.Create(models.AutoRent{
			AutoID:    "TestGetCurrentCommission",
			StartDate: testday.AddDate(0, 0, -10),
			EndDate:   testday.AddDate(0, 0, 7),
		})

		want := (10 * 200) + (4 * 200 * 20 / 100) + 200
		comm, ins, err := svc.GetCurrentCommission("TestGetCurrentCommission", testday)
		if err != nil {
			t.Error(err)
		}
		if comm != want {
			t.Errorf("want %d, got %d", want, comm)
		}
		if ins != 200 {
			t.Errorf("want %d, got %d", 200, ins)
		}
	}
	{
		testday := time.Date(2023, time.September, 15, 1, 2, 3, 4, time.UTC)
		db.Create(&models.Auto{ID: "TestGetCurrentCommission1", Type: "TestGetCurrentCommission"})
		db.Create(models.AutoRent{
			AutoID:    "TestGetCurrentCommission1",
			StartDate: testday,
			EndDate:   testday.AddDate(0, 0, 11),
		})

		want := 1*200 + 200
		comm, ins, err := svc.GetCurrentCommission("TestGetCurrentCommission1", testday)
		if err != nil {
			t.Error(err)
		}
		if comm != want {
			t.Errorf("want %d, got %d", want, comm)
		}
		if ins != 200 {
			t.Errorf("want %d, got %d", 200, ins)
		}
	}
	db.Delete(&models.AutoRent{}, "auto_id = ?", "TestGetCurrentCommission")
	db.Delete(&models.Auto{}, "type = ?", "TestGetCurrentCommission")
	db.Delete(&models.Commission{}, "auto_type = ?", "TestGetCurrentCommission")
	db.Delete(&models.Commission{}, "auto_type = ?", "TestGetCurrentCommission1")
	db.Delete(&models.AutoRent{}, "auto_id = ?", "TestGetCurrentCommission1")
	db.Delete(&models.Auto{}, "type = ?", "TestGetCurrentCommission1")
	db.Delete(&models.Commission{}, "auto_type = ?", "TestGetCurrentCommission1")
	db.Delete(&models.AutoType{}, "id = ?", "TestGetCurrentCommission")
	db.Delete(&models.AutoType{}, "id = ?", "TestGetCurrentCommission1")
}
