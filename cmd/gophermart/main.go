package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/with0p/gophermart/internal/config"
	"github.com/with0p/gophermart/internal/handlers"
	"github.com/with0p/gophermart/internal/logger"
	"github.com/with0p/gophermart/internal/models"
	"github.com/with0p/gophermart/internal/service"
	"github.com/with0p/gophermart/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	config := config.GetConfig()

	db, dbErr := sql.Open("pgx", config.DataBaseAddress)
	if dbErr != nil {
		logger.Error(dbErr)
		return
	}
	defer db.Close()

	ctx, cancelInitDB := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelInitDB()

	storage, err := storage.NewStorageDB(ctx, db)
	if err != nil {
		logger.Error(err)
		return
	}

	queue := make(chan models.OrderID, 10)
	service := service.NewServiceGophermart(storage)
	handler := handlers.NewHandlerUserAPI(&service, queue)
	router := handler.GetHandlerUserAPIRouter()
	server := &http.Server{Addr: config.BaseURL, Handler: router}

	//run accrual
	var accrualCmd *exec.Cmd
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		accrualCmd = startAccrualService(config.AccrualURL)
	}()

	//start processing routine
	go service.ProcessOrders(queue, config.AccrualURL)

	//run periodic queue feed to process unfinished orders
	go func() {
		interval := time.Minute
		for {
			time.Sleep(interval)
			service.FeedQueue(queue)
		}
	}()

	//run gophermart
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Error(err)
			return
		}
	}()

	//server shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("Starting to shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error(err)
	}

	logger.Info("Server is stopped")

	if accrualCmd != nil {
		if err := accrualCmd.Process.Signal(syscall.SIGTERM); err != nil {
			logger.Error(err)
		}

		if err := accrualCmd.Wait(); err != nil {
			logger.Error(err)
		}
		logger.Info("Accrual service stopped.")
	}

	close(queue)
	wg.Wait()

	logger.Info("All services stopped.")
}

func startAccrualService(url string) *exec.Cmd {
	cmd := exec.Command("./accrual_darwin_arm64", "-a", url)
	cmd.Dir = "cmd/accrual"

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logger.Error(err)
		return nil
	}

	logger.Info("Accrual is running")
	return cmd
}
