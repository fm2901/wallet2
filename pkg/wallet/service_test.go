package wallet_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/fm2901/wallet/pkg/types"
	"github.com/fm2901/wallet/pkg/wallet"
	"github.com/google/uuid"
)

type testService struct {
	*wallet.Service
}

type testAccount struct {
	phone types.Phone
	balance types.Money
	payments []struct {
		amount types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = testAccount {
	phone: "+992926409003",
	balance: 10_000_00,
	payments: []struct {
		amount types.Money
		category types.PaymentCategory 
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}

	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposity account, error = %v", err)
	}

	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}
	}

	return account, payments, nil
}

func newTestService() *testService {
	return &testService{Service: &wallet.Service{}}
}

func (s * testService) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {
	account, err := s.RegisterAccount(phone)
	if err != nil {
		return nil, fmt.Errorf("can't register account, error = %v", err)
	}

	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, fmt.Errorf("can't deposit account, error = %v", err)
	}

	return account, nil
}

func TestService_FindPaymentByID_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}

	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID(): wrong payment returned = %v", err)
		return
	}
}

func TestService_FindPaymentByID_fail(t *testing.T) {
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}
	
	if err != wallet.ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}
}

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
	s := newTestService()

	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}
	
	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't find payment by id, error = %v", err)
		return
	}
	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject(): status didn't cnahged, payment = %v", savedPayment)
		return
	}

	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): can't find account by id, error = %v", err)
		return
	}
	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject(): balance didn't changed, account = %v", savedAccount)
		return
	}
	
}

func TestService_Repeat_success(t *testing.T) {
	s := newTestService()

	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	_, err = s.Repeat(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't repeat, payment = %v", payment)
		return
	}
}

func TestService_FavoritePayment_success(t *testing.T) {
	s := newTestService()

	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	_, err = s.FavoritePayment(payment.ID, "first favorite")
	if err != nil {
		t.Errorf("FavoritePayment(): error = %v", err)
		return
	}
}

func TestService_PayFromFavorite_success(t *testing.T) {
	s := newTestService()

	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	favorite, err := s.FavoritePayment(payment.ID, "first favorite")
	if err != nil {
		t.Errorf("FavoritePayment(): error = %v", err)
		return
	}

	payment, err = s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("PayFromFavorite(): error = %v", err)
		return
	}
}