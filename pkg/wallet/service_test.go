package wallet_test

import (
	"testing"
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