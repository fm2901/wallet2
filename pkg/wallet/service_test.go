package wallet_test

import (
	"testing"
	"github.com/fm2901/wallet/pkg/types"
	"github.com/fm2901/wallet/pkg/wallet"
)
func TestFindAccountByID_empty(t *testing.T) {
	svc := &wallet.Service{}
	result, err := svc.FindAccountByID(1)
	if err != wallet.ErrAccountNotFound || result != nil {
		t.Error("Тест empty не прошел")
	}
}

func TestFindAccountByID_notEmpty(t *testing.T) {
	svc := &wallet.Service{}
	result, err := svc.RegisterAccount("+992000000001")
	result, err = svc.FindAccountByID(3)
	if err != wallet.ErrAccountNotFound || result != nil {
		t.Error("Тест notEmpty не прошел")
	}
}

func TestReject_empty(t *testing.T) {
	svc := &wallet.Service{}
	err := svc.Reject("1")
	if err != wallet.ErrPaymentNotFound{
		t.Error("Тест Reject_empty не прошел")
	}
}

func TestService_Reject_success(t *testing.T) {
	svc := &wallet.Service{}
	phone := types.Phone("+992926409000")
	account, err := svc.RegisterAccount(phone)
	if err != nil {
		t.Errorf("Reject(): can't register account, error = %v", err)
		return
	}

	err = svc.Deposit(account.ID, 100)
	if err != nil {
		t.Errorf("Reject(): can't deposit, error = %v", err)
		return
	}

	payment, err := svc.Pay(account.ID, 10, "car")
	if err != nil {
		t.Errorf("Reject(): can't create payment, error = %v", err)
		return
	}

	err = svc.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't reject payment, error = %v", err)
		return
	}
}

