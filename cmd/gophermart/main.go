package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/with0p/gophermart/internal/config"
	"github.com/with0p/gophermart/internal/handlers"
	"github.com/with0p/gophermart/internal/models"
	"github.com/with0p/gophermart/internal/service"
	"github.com/with0p/gophermart/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	config := config.GetConfig()

	db, dbErr := sql.Open("pgx", config.DataBaseAddress)
	if dbErr != nil {
		fmt.Println(dbErr.Error())
	}
	defer db.Close()

	ctx, cancelInitDB := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelInitDB()

	storage, err := storage.NewStorageDB(ctx, db)
	if err != nil {
		fmt.Println(err.Error())
	}

	queue := make(chan models.OrderID, 10)
	service := service.NewServiceGophermart(storage)
	handler := handlers.NewHandlerUserAPI(&service, queue)
	router := handler.GetHandlerUserAPIRouter()
	server := &http.Server{Addr: config.BaseURL, Handler: router}

	//run accrual
	var wg sync.WaitGroup
	wg.Add(1)
	var accrualCmd *exec.Cmd
	go func() {
		defer wg.Done()
		accrualCmd = startAccrualService(config.AccrualURL)
	}()

	// //start processing routine
	go service.ProcessOrders(queue)

	// //run periodic queue feed to process unfinished orders
	go func() {
		interval := time.Minute
		for {
			time.Sleep(interval)
			service.FeedQueue(queue)
		}
	}()

	// //run gophermart
	go func() {
		serverErr := server.ListenAndServe()
		if err != nil {
			fmt.Println(serverErr.Error())
			return
		}
	}()

	//server shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	fmt.Println("Starting to shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Server is stopped")

	if accrualCmd != nil {
		if err := accrualCmd.Process.Signal(syscall.SIGTERM); err != nil {
			fmt.Println(err.Error())
		}

		if err := accrualCmd.Wait(); err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("Accrual service stopped.")
	}

	close(queue)
	wg.Wait()

	fmt.Println("All services stopped.")
}

func startAccrualService(url string) *exec.Cmd {
	cmd := exec.Command("./accrual_darwin_arm64", "-a", url)
	cmd.Dir = "cmd/accrual"

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println(err.Error())
		return nil
	}

	fmt.Println("Accrual is running")
	return cmd
}
