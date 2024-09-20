package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/theplant/luhn"
	customerror "github.com/with0p/gophermart/internal/custom-error"
	"github.com/with0p/gophermart/internal/logger"
	"github.com/with0p/gophermart/internal/models"
	"github.com/with0p/gophermart/internal/storage"
	"github.com/with0p/gophermart/internal/utils"
)

type ServiceGophermart struct {
	storage storage.Storage
}

func NewServiceGophermart(currentStorage storage.Storage) ServiceGophermart {
	return ServiceGophermart{
		storage: currentStorage,
	}
}

func (s *ServiceGophermart) RegisterUser(ctx context.Context, login string, password string) error {
	return s.storage.CreateUser(ctx, login, utils.HashPassword(password))
}

func (s *ServiceGophermart) AuthenticateUser(ctx context.Context, login string, password string) error {
	return s.storage.ValidateUser(ctx, login, utils.HashPassword(password))
}

func (s *ServiceGophermart) AddOrder(ctx context.Context, login string, orderID models.OrderID) error {
	orderIDInt, errInt := strconv.ParseInt(string(orderID), 10, 64)
	if errInt != nil || !luhn.Valid(int(orderIDInt)) {
		return customerror.ErrWrongOrderFormat
	}

	userID, errUser := s.storage.GetUserID(ctx, login)
	if errUser != nil {
		return errUser
	}

	order, errOrder := s.storage.GetOrder(ctx, orderID)
	if errOrder != nil {
		return errOrder
	}

	if order == nil {
		return s.storage.AddOrder(ctx, userID, models.StatusNew, orderID)
	}

	if order.UserID != userID {
		return customerror.ErrAnotherUserOrder
	}

	return customerror.ErrAlreadyAdded
}

func (s *ServiceGophermart) GetUserOrders(ctx context.Context, login string) ([]models.Order, error) {
	userID, errUser := s.storage.GetUserID(ctx, login)
	if errUser != nil {
		return nil, errUser
	}

	orders, err := s.storage.GetUserOrders(ctx, userID)
	var ordersFormatted = make([]models.Order, len(orders))
	for i, order := range orders {
		orderUploadedParsed, err := time.Parse("2006-01-02 15:04:05.999999-07", order.UploadDate)
		if err != nil {
			logger.Error(err)
			continue
		}
		loc, errLoc := time.LoadLocation("Europe/Moscow")
		if errLoc != nil {
			logger.Error(err)
			continue
		}
		orderUploadedParsed = orderUploadedParsed.In(loc)
		ordersFormatted[i] = order
		ordersFormatted[i].UploadDate = orderUploadedParsed.Format(time.RFC3339)
	}

	return ordersFormatted, err
}

func (s *ServiceGophermart) FeedQueue(queue chan models.OrderID) {
	logger.Info("feedQueue")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ordersToProcess, err := s.storage.GetUnfinishedOrderIDs(ctx)
	if err != nil {
		logger.Error(err)
		return
	}

	for _, order := range ordersToProcess {
		queue <- order
	}
}

func (s *ServiceGophermart) ProcessOrders(queue chan models.OrderID, accrualAddr string) {
	logger.Info("ProcessOrders")
	ctx := context.Background()

	var wg sync.WaitGroup

	numWorkers := 3
	semaphore := utils.NewSemaphore(numWorkers)
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, queue, &wg, s, semaphore, accrualAddr)
	}

	wg.Wait()
	logger.Info("All workers finished processing.")

}

func worker(ctx context.Context, jobs chan models.OrderID, wg *sync.WaitGroup, s *ServiceGophermart, sem *utils.Semaphore, accrualAddr string) {
	defer wg.Done()
	logger.Info("worker")
	for orderID := range jobs {
		sem.Acquire()
		orderData, err := getOrderDataFromAccrual(ctx, orderID, accrualAddr)
		if err != nil {
			if errors.Is(err, customerror.ErrTooManyRequests) {
				sem.Acquire()
				logger.Info("Too many requests, waiting 60s")
				time.Sleep(60 * time.Second)
				sem.Release()
				logger.Info("Going back to processing")
				jobs <- orderID
			}
			logger.Error(err)
			continue
		}

		errOrd := s.storage.UpdateOrder(ctx, orderID, models.OrderStatus(orderData.Status), float32(orderData.Accrual))
		if errOrd != nil {
			logger.Error(errOrd)
		}
		sem.Release()
	}
}

func getOrderDataFromAccrual(ctx context.Context, orderID models.OrderID, accrualAddr string) (*models.OrderExternalData, error) {
	client := &http.Client{
		Timeout: 2 * time.Minute,
	}

	url := fmt.Sprintf("%s/api/orders/%s", accrualAddr, orderID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, customerror.ErrTooManyRequests
		}
		return nil, fmt.Errorf("not accepted, status %d", resp.StatusCode)
	}

	var orderData models.OrderExternalData
	err = json.Unmarshal(body, &orderData)
	if err != nil {
		return nil, err
	}

	return &orderData, nil
}

func (s *ServiceGophermart) MakeWithdrawal(ctx context.Context, login string, orderID models.OrderID, amount float32) error {
	orderIDInt, errInt := strconv.ParseInt(string(orderID), 10, 64)
	if errInt != nil || !luhn.Valid(int(orderIDInt)) {
		return customerror.ErrWrongOrderFormat
	}

	userID, err := s.storage.GetUserID(ctx, login)
	if err != nil {
		return err
	}

	return s.storage.AddWithdrawal(ctx, userID, orderID, amount)
}

func (s *ServiceGophermart) GetUserBalance(ctx context.Context, login string) (*models.Balance, error) {
	userID, err := s.storage.GetUserID(ctx, login)
	if err != nil {
		return nil, err
	}

	accrualSum, err := s.storage.GetUserAccrualBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	withdrawSum, err := s.storage.GetUserWithdrawalSum(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.Balance{
		Current:   accrualSum - withdrawSum,
		Withdrawn: withdrawSum,
	}, nil
}

func (s *ServiceGophermart) GetUserWithdrawals(ctx context.Context, login string) ([]models.Withdrawal, error) {
	userID, errUser := s.storage.GetUserID(ctx, login)
	if errUser != nil {
		return nil, errUser
	}

	withdrawals, err := s.storage.GetUserWithdrawals(ctx, userID)
	var withdrawalsFormatted = make([]models.Withdrawal, len(withdrawals))
	for i, withdrawal := range withdrawals {
		withdrawalUploadedParsed, err := time.Parse("2006-01-02 15:04:05.999999-07", withdrawal.ProcessedAt)
		if err != nil {
			continue
		}
		loc, errLoc := time.LoadLocation("Europe/Moscow")
		if errLoc != nil {
			continue
		}
		withdrawalUploadedParsed = withdrawalUploadedParsed.In(loc)
		withdrawalsFormatted[i] = withdrawal
		withdrawalsFormatted[i].ProcessedAt = withdrawalUploadedParsed.Format(time.RFC3339)
	}

	return withdrawalsFormatted, err
}
