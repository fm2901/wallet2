package wallet

import (
	"fmt"
	"testing"
	"reflect"
	"github.com/fm2901/wallet/pkg/types"
)

type testService struct {
	*Service
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = testAccount{
	phone:   "992000000001",
	balance: 10_000_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
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

func TestService_RegisterAccount_success(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9920000001")

	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", account)
	}
}

func TestService_FindAccoundByIdmethod_notFound(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9920000001")

	account, err := svc.FindAccountByID(2)
	if err == nil {
		t.Errorf("\ngot > %v \nwant > nil", account)
	}
}

func TestDeposit(t *testing.T) {
	svc := Service{}

	svc.RegisterAccount("+992000000001")

	err := svc.Deposit(1, 100_00)
	if err != nil {
		t.Error("something wrong while paying")
	}

	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", account)
	}
}

func TestService_Reject_success(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9920000001")

	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	err = svc.Deposit(account.ID, 1000_00)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	payment, err := svc.Pay(account.ID, 100_00, "auto")
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	pay, err := svc.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	err = svc.Reject(pay.ID)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}
}

func TestService_Reject_fail(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9920000001")

	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	err = svc.Deposit(account.ID, 1000_00)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	payment, err := svc.Pay(account.ID, 100_00, "auto")
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	pay, err := svc.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	editPayID := pay.ID + "mr.virus :)"
	err = svc.Reject(editPayID)
	if err == nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}
}

func TestService_Repeat_success(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9920000001")

	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	err = svc.Deposit(account.ID, 1000_00)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	payment, err := svc.Pay(account.ID, 100_00, "auto")
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	pay, err := svc.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	pay, err = svc.Repeat(pay.ID)
	if err != nil {
		t.Errorf("Repeat(): Error(): can't pay for an account(%v): %v", pay.ID, err)
	}
}

func TestService_Favorite_success_user(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
		t.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	payment, err := svc.Pay(account.ID, 10_00, "auto")
	if err != nil {
		t.Errorf("Pay() Error() can't pay for an account(%v): %v", account, err)
	}

	favorite, err := svc.FavoritePayment(payment.ID, "megafon")
	if err != nil {
		t.Errorf("FavoritePayment() Error() can't for an favorite(%v): %v", favorite, err)
	}

	paymentFavorite, err := svc.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("PayFromFavorite() Error() can't for an favorite(%v): %v", paymentFavorite, err)
	}
}

func TestExportToFile_success(t *testing.T) {
	svc := Service{}
	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}
	file := svc.ExportToFile("message.txt")
	if file != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}
}

func TestImportFromFile_success(t *testing.T) {
	svc := Service{}
	file := svc.ImportFromFile("message.txt")
	if file != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", file)
	}
}

func TestExport_success(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
		t.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	payment, err := svc.Pay(account.ID, 10_00, "auto")
	if err != nil {
		t.Errorf("Pay() Error() can't pay for an account(%v): %v", account, err)
	}

	favorite, err := svc.FavoritePayment(payment.ID, "megafon")
	if err != nil {
		t.Errorf("FavoritePayment() Error() can't for an favorite(%v): %v", favorite, err)
	}

	paymentFavorite, err := svc.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("PayFromFavorite() Error() can't for an favorite(%v): %v", paymentFavorite, err)
	}

	file := svc.Export("./data")
	if file != nil {
		t.Errorf("PayFromFavorite() Error() can't for an favorite(%v): %v", paymentFavorite, err)
	}
}

func TestImport(t *testing.T) {
	svc := Service{}
	file := svc.Export("./data")
	if file != nil {
		t.Errorf("PayFromFavorite() Error() can't for an favorite: %v", file)
	}
}

func TestExportAccountHistory_success(t *testing.T) {
	svc := Service{}
	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
		t.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	payment, err := svc.Pay(account.ID, 10_00, "auto")
	if err != nil {
		t.Errorf("Pay() Error() can't pay for an account(%v): %v", account, err)
	}
	file, err := svc.ExportAccountHistory(account.ID)
	if err != nil {
		t.Errorf("Pay() Error() can't pay for an account(%v): %v", payment, file)
	}
}

func TestHistoryToFiles(t *testing.T) {
	svc := Service{}
	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
		t.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	payment, err := svc.Pay(account.ID, 10_00, "auto")
	if err != nil {
		t.Errorf("Pay() Error() can't pay for an account(%v): %v", account, err)
	}
	file, err := svc.ExportAccountHistory(account.ID)
	if err != nil {
		t.Errorf("Pay() Error() can't pay for an account(%v): %v", payment, err)
	}
	files := svc.HistoryToFiles(file, "./data", 1)
	if files != nil {
		t.Errorf("Pay() Error() can't pay for an account(%v): %v", payment, err)
	}
}
func BenchmarkSumPayments(b *testing.B) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		b.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
		b.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	payment, err := svc.Pay(account.ID, 10_00, "auto")
	if err != nil {
		b.Errorf("Pay() Error() can't pay for an account(%v): %v", account, err)
	}
	favorite, err := svc.FavoritePayment(payment.ID, "megafon")
	if err != nil {
		b.Errorf("FavoritePayment() Error() can't for an favorite(%v): %v", favorite, err)
	}

	paymentFavorite, err := svc.PayFromFavorite(favorite.ID)
	if err != nil {
		b.Errorf("PayFromFavorite() Error() can't for an favorite(%v): %v", paymentFavorite, err)
	}

	want := types.Money(2000)
	for i := 0; i < b.N; i++ {
		result := svc.SumPayments(1)
		if result != want {
			b.Fatalf("invalid result, got %v, want %v", result, want)
		}
	}
}

func BenchmarkFilterPayments(b *testing.B) {
	s := newTestService()
	account, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		b.Error(err)
		return
	}

	want := []types.Payment{}
	for _, pay := range s.payments {
		want = append(want, *pay)
	}

	for i := 0; i < b.N; i++ {
		result, _ := s.FilterPayments(account.ID, 3)

		b.StopTimer()
		if !reflect.DeepEqual(result, want) {
			b.Fatalf("invalid result got: %v; want: %v", result, want)
		}
		b.StartTimer()
	}
}
