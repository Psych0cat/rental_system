package main

import (
	"net/http"
	"os"
	"time"

	"car-rental/internal/controller"
	"car-rental/internal/models"
	"car-rental/internal/repository"
	"car-rental/internal/router"
	"car-rental/internal/service"
	"car-rental/internal/utils"
)

func main() {
	dsn := utils.GetDsnFromEnv()
	db, err := models.ConnectDatabase(dsn)

	autoRepository := repository.NewAutoRepositoryImpl(db)
	commissionRepository := repository.NewCommissionRepositoryImpl(db)
	rentalRepository := repository.NewRentalRepositoryImpl(db)

	rentalService := service.NewRentalServiceImpl(autoRepository, rentalRepository, commissionRepository)
	rentalController := controller.NewRentalController(rentalService)
	routes := router.NewRouter(*rentalController)

	port := os.Getenv("port")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        routes,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
