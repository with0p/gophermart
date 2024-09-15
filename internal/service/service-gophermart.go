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
			continue
		}
		loc, errLoc := time.LoadLocation("Europe/Moscow")
		if errLoc != nil {
			continue
		}
		orderUploadedParsed = orderUploadedParsed.In(loc)
		ordersFormatted[i] = order
		ordersFormatted[i].UploadDate = orderUploadedParsed.Format(time.RFC3339)
	}

	return ordersFormatted, err
}

func (s *ServiceGophermart) FeedQueue(queue chan models.OrderID) {
	fmt.Println("feedQueue")
	ctx1, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ordersToProcess, err := s.storage.GetUnfinishedOrderIDs(ctx1)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, order := range ordersToProcess {
		queue <- order
	}
}

func (s *ServiceGophermart) ProcessOrders(queue chan models.OrderID) {
	fmt.Println("ProcessOrders")
	ctx1, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	numWorkers := 3
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(queue, &wg, s, ctx1)
	}

	wg.Wait()
	fmt.Println("All workers finished processing.")

}

func worker(jobs <-chan models.OrderID, wg *sync.WaitGroup, s *ServiceGophermart, ctx context.Context) {
	defer wg.Done()
	fmt.Println("worker")
	for orderID := range jobs {
		orderData, err := getOrderDataFromAccrual(orderID)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		errOrd := s.storage.UpdateOrder(ctx, orderID, models.OrderStatus(orderData.Status), int(orderData.Accrual))
		if errOrd != nil {
			fmt.Println(errOrd.Error())
		}
	}
}

func getOrderDataFromAccrual(orderID models.OrderID) (*models.OrderExternalData, error) {
	url := fmt.Sprintf("http://localhost:8081/api/orders/%s", orderID)
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	fmt.Println(resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("not accepted")
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	var orderData models.OrderExternalData
	err = json.Unmarshal(body, &orderData)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &orderData, nil
}

func (s *ServiceGophermart) MakeWithdrawal(ctx context.Context, login string, orderID models.OrderID, amount int) error {
	orderIDInt, errInt := strconv.ParseInt(string(orderID), 10, 64)
	if errInt != nil || !luhn.Valid(int(orderIDInt)) {
		return customerror.ErrWrongOrderFormat
	}

	userID, err := s.storage.GetUserID(ctx, login)
	if err != nil {
		return err
	}

	accrualBalance, err := s.storage.GetUserAccrualBalance(ctx, userID)
	if err != nil {
		return err
	}

	if accrualBalance < amount {
		return customerror.ErrInsufficientBalance
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
		Current:   accrualSum,
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
