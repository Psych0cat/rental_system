package controller

import (
	"errors"
	"net/http"
	"time"

	"car-rental/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const internalError = "internal server error"

type RentalController struct {
	rentalService service.RentalService
}

func NewRentalController(rentalService service.RentalService) *RentalController {
	return &RentalController{rentalService: rentalService}
}

func (r RentalController) GetAvailableByType(ctx *gin.Context) {
	autos, err := r.rentalService.GetAvailableAutoByType(ctx.Params.ByName("type"))
	if err != nil {
		if err.Error() == service.NotFoundError {
			ctx.JSON(404, err.Error())
			return
		}
		ctx.JSON(500, internalError)
		return
	}
	ctx.JSON(200, autos)
}

func (r RentalController) BindAuto(ctx *gin.Context) {
	var input struct {
		AutoId string `json:"auto_id"`
		Days   int    `json:"days"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Days <= 0 {
		ctx.JSON(403, "days must be positive")
		return
	}
	err := r.rentalService.BindAuto(input.AutoId, input.Days)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == service.NotFoundError {
			ctx.JSON(404, "auto is not available")
			return
		} else if err.Error() == service.AlreadyRentedError {
			ctx.JSON(400, err.Error())
			return
		} else if err.Error() == service.ThresholdValidationError {
			ctx.JSON(403, err.Error())
			return
		}
		ctx.JSON(500, errors.New(internalError))
		return

	}
	ctx.JSON(200, "ok")
}
func (r RentalController) ReleaseAuto(ctx *gin.Context) {
	checkout, err := r.rentalService.ReleaseAuto(ctx.Params.ByName("auto_id"), time.Now())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(404, err.Error())
			return
		} else {
			ctx.JSON(500, errors.New(internalError))
			return
		}
	}
	if err != nil {
		ctx.JSON(500, err.Error())
		return
	}
	ctx.JSON(200, gin.H{
		"checkout": checkout,
	})
}

func (r RentalController) GetCurrentCommission(ctx *gin.Context) {
	commission, insurance, err := r.rentalService.GetCurrentCommission(ctx.Params.ByName("auto_id"), time.Now())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(404, "rent not found")
			return
		} else {
			ctx.JSON(500, errors.New("internal server error"))
			return
		}
	}
	ctx.JSON(200, gin.H{
		"commission": commission,
		"insurance":  insurance,
	})
}
