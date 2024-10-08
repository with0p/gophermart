package customerror

import "errors"

var ErrUniqueKeyConstrantViolation = errors.New("unique key violation")
var ErrNoSuchUser = errors.New("no such user")
var ErrAnotherUserOrder = errors.New("another user's order")
var ErrAlreadyAdded = errors.New("order already added by this user")
var ErrWrongOrderFormat = errors.New("wrong order format")
var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrTooManyRequests = errors.New("to many requests")
