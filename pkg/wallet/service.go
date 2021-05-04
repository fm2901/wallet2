package wallet

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/fm2901/wallet/pkg/types"
	"github.com/google/uuid"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountmustBePositive = errors.New("amount must be a greater then zero")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("not enough balance")
var ErrPaymentNotFound = errors.New("payment not found by id")
var ErrFavoriteNotFound = errors.New("favorite not found")

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}

	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountmustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return ErrAccountNotFound
	}

	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountmustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			return acc, nil
		}
	}
	return nil, ErrAccountNotFound
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}
	return nil, ErrPaymentNotFound
}

func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}

	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}

	account.Balance += payment.Amount
	payment.Status = types.PaymentStatusFail

	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	payment, err = s.Pay(payment.AccountID, payment.Amount, payment.Category)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	favorite := &types.Favorite{
		ID:        uuid.New().String(),
		AccountID: payment.AccountID,
		Amount:    payment.Amount,
		Category:  payment.Category,
		Name:      name,
	}
	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}
	return nil, ErrFavoriteNotFound
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}

	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) ExportToFile(path string) error {
	accountStr := ""
	file, err := os.Create(path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	for _, acc := range s.accounts {
		id := strconv.Itoa(int(acc.ID))
		phone := string(acc.Phone)
		balance := strconv.Itoa(int(acc.Balance))
		accountStr += id + ";" + phone + ";" + balance + "|"
	}
	accountStr = accountStr[:len(accountStr)-1]
	_, err = file.Write([]byte(accountStr))
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func (s *Service) ImportFromFile(path string) error {
	content := make([]byte, 0)
	buf := make([]byte, 4)
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			content = append(content, buf[:read]...)
			break
		}

		if err != nil {
			log.Print(err)
			return err
		}
		content = append(content, buf[:read]...)
	}

	data := string(content)

	rows := strings.Split(data, "|")
	for _, row := range rows {
		cols := strings.Split(row, ";")
		id, _ := strconv.ParseInt(cols[0], 10, 64)
		phone := types.Phone(cols[1])
		balance, _ := strconv.ParseInt(cols[2], 10, 64)

		account := &types.Account{
			ID:      id,
			Phone:   phone,
			Balance: types.Money(balance),
		}
		s.accounts = append(s.accounts, account)
	}

	return nil
}

func (s *Service) Export(dir string) error {
	if len(s.accounts) > 0 {
		file, err := os.OpenFile(dir+"/accounts.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		defer func() {
			if cerr := file.Close(); cerr != nil {
				if err != nil {
					err = cerr
					log.Print(err)
				}
			}
		}()
		fileStr := ""
		for _, account := range s.accounts {
			fileStr += fmt.Sprint(account.ID) + ";" + string(account.Phone) + ";" + fmt.Sprint(account.Balance) + "\n"
		}
		file.WriteString(fileStr[:len(fileStr)-1])
	}
	if len(s.payments) > 0 {
		file, err := os.OpenFile(dir+"/payments.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		defer func() {
			if cerr := file.Close(); cerr != nil {
				if err != nil {
					err = cerr
					log.Print(err)
				}
			}
		}()
		fileStr := ""
		for _, payment := range s.payments {
			fileStr += fmt.Sprint(payment.ID) + ";" + fmt.Sprint(payment.AccountID) + ";" + fmt.Sprint(payment.Amount) + ";" + fmt.Sprint(payment.Category) + ";" + fmt.Sprint(payment.Status) + "\n"
		}
		file.WriteString(fileStr[:len(fileStr)-1])
	}
	if len(s.favorites) > 0 {
		file, err := os.OpenFile(dir+"/favorites.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		defer func() {
			if cerr := file.Close(); cerr != nil {
				if err != nil {
					err = cerr
					log.Print(err)
				}
			}
		}()

		fileStr := ""
		for _, favorite := range s.favorites {
			fileStr += fmt.Sprint(favorite.ID) + ";" + fmt.Sprint(favorite.AccountID) + ";" + favorite.Name + ";" + fmt.Sprint(favorite.Amount) + ";" + fmt.Sprint(favorite.Category) + "\n"
		}
		file.WriteString(fileStr[:len(fileStr)-1])
	}
	return nil
}

func (s *Service) Import(dir string) error {
	_, err := os.Stat(dir + "/accounts.dump")
	if err == nil {
		content, err := ioutil.ReadFile(dir + "/accounts.dump")
		if err != nil {
			return err
		}
		rows := strings.Split(string(content), "\n")
		for _, row := range rows {
			cols := strings.Split(row, ";")

			id, err := strconv.ParseInt(cols[0], 10, 64)
			if err != nil {
				return err
			}
			balance, err := strconv.ParseInt(cols[2], 10, 64)
			if err != nil {
				return err
			}
			flag := true
			for _, v := range s.accounts {
				if v.ID == id {
					flag = false
				}
			}
			if flag {
				account := &types.Account{
					ID:      id,
					Phone:   types.Phone(cols[1]),
					Balance: types.Money(balance),
				}
				s.accounts = append(s.accounts, account)
			}
		}
	}
	_, err = os.Stat(dir + "/payments.dump")
	if err == nil {
		content, err := ioutil.ReadFile(dir + "/payments.dump")
		if err != nil {
			return err
		}
		rows := strings.Split(string(content), "\n")
		for _, row := range rows {
			cols := strings.Split(row, ";")

			id := cols[0]
			if err != nil {
				return err
			}
			accountID, err := strconv.ParseInt(cols[1], 10, 64)
			if err != nil {
				return err
			}
			amount, err := strconv.ParseInt(cols[2], 10, 64)
			if err != nil {
				return err
			}
			flag := true
			for _, v := range s.payments {
				if v.ID == id {
					flag = false
				}
			}
			if flag {
				data := &types.Payment{
					ID:        id,
					AccountID: accountID,
					Amount:    types.Money(amount),
					Category:  types.PaymentCategory(cols[3]),
					Status:    types.PaymentStatus(cols[4]),
				}
				s.payments = append(s.payments, data)
			}
		}

	}

	_, err = os.Stat(dir + "/favorites.dump")
	if err == nil {
		content, err := ioutil.ReadFile(dir + "/favorites.dump")
		if err != nil {
			return err
		}
		rows := strings.Split(string(content), "\n")
		for _, row := range rows {
			cols := strings.Split(row, ";")

			id := cols[0]
			if err != nil {
				return err
			}
			accountID, err := strconv.ParseInt(cols[1], 10, 64)
			if err != nil {
				return err
			}
			amount, err := strconv.ParseInt(cols[3], 10, 64)
			if err != nil {
				return err
			}
			flag := true
			for _, v := range s.favorites {
				if v.ID == id {
					flag = false
				}
			}
			if flag {
				data := &types.Favorite{
					ID:        id,
					AccountID: accountID,
					Name:      cols[2],
					Amount:    types.Money(amount),
					Category:  types.PaymentCategory(cols[4]),
				}
				s.favorites = append(s.favorites, data)
			}
		}
	}
	return nil
}

func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {
	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return nil, err
	}

	history := []types.Payment{}
	for _, payment := range s.payments {
		if payment.AccountID == account.ID {
			curPayment := types.Payment{
				ID:        payment.ID,
				AccountID: payment.AccountID,
				Amount:    payment.Amount,
				Category:  payment.Category,
				Status:    payment.Status,
			}
			history = append(history, curPayment)
		}
	}
	return history, nil
}

func HistoryToFile(payments []types.Payment, filename string) error {
	if len(payments) < 1 {
		return nil
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	defer func() {
		if cerr := file.Close(); cerr != nil {
			if err != nil {
				err = cerr
				log.Print(err)
			}
		}
	}()
	fileStr := ""
	for _, payment := range payments {
		fileStr += fmt.Sprint(payment.ID) + ";" + fmt.Sprint(payment.AccountID) + ";" + fmt.Sprint(payment.Amount) + ";" + fmt.Sprint(payment.Category) + ";" + fmt.Sprint(payment.Status) + "\n"
	}
	file.WriteString(fileStr[:len(fileStr)-1])
	return nil
}

func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {
	if len(payments) < 1 {
		return nil
	}
	if len(payments) <= records {
		HistoryToFile(payments, dir+"payments.dump")
	} else {
		counter := 1
		fIndex := 0
		lIndex := records
		for {
			HistoryToFile(payments[fIndex:lIndex], dir+"payments"+fmt.Sprint(counter)+".dump")
			fIndex += records
			lIndex += records
			if lIndex >= len(payments) {
				if counter*records < len(payments) {
					lIndex = len(payments) - counter*records
					HistoryToFile(payments[:lIndex], dir+"payments"+fmt.Sprint(counter+1)+".dump")
				}
				break
			}
			counter++
		}
	}
	return nil
}

func (s *Service) SumPayments(goroutines int) types.Money {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	sum := int64(0)
	kol := 0
	i := 0
	if goroutines == 0 {
		kol = len(s.payments)
	} else {
		kol = int(len(s.payments) / goroutines)
	}
	for i = 0; i < goroutines-1; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			val := int64(0)
			payments := s.payments[index*kol : (index+1)*kol]
			for _, payment := range payments {
				val += int64(payment.Amount)
			}
			mu.Lock()
			sum += val
			mu.Unlock()

		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		val := int64(0)
		payments := s.payments[i*kol:]
		for _, payment := range payments {
			val += int64(payment.Amount)
		}
		mu.Lock()
		sum += val
		mu.Unlock()

	}()
	wg.Wait()
	return types.Money(sum)
}

func (s *Service) FilterPayments(accountID int64, goroutines int) ([]types.Payment, error) {

	if goroutines == 0 || goroutines == 1 {
		payments, err := s.ExportAccountHistory(accountID)
		if err != nil {
			return nil, err
		}
		return payments, nil
	}

	_, err := s.FindAccountByID(accountID)
	if err != nil {
		return nil, err
	}

	payments := []types.Payment{}
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	paymentsOnGoroutine := len(s.payments) / goroutines
	count := 0
	for i := 0; i < len(s.payments); i++ {
		if i == len(s.payments)-1 {
			wg.Add(1)
			go func(count, i int) {
				defer wg.Done()
				tmp := []types.Payment{}
				for _, payment := range s.payments[count:] {
					if payment.AccountID == accountID {
						tmp = append(tmp, *payment)
					}
				}
				mu.Lock()
				defer mu.Unlock()
				payments = append(payments, tmp...)
			}(count, i)
		}

		if i-count == paymentsOnGoroutine {
			wg.Add(1)
			go func(count, i int) {
				defer wg.Done()
				tmp := []types.Payment{}
				for _, payment := range s.payments[count:i] {
					if payment.AccountID == accountID {
						tmp = append(tmp, *payment)
					}
				}
				mu.Lock()
				defer mu.Unlock()
				payments = append(payments, tmp...)
			}(count, i)
			count += paymentsOnGoroutine
		}
	}
	wg.Wait()
	return payments, nil
}

func (s *Service) SumPaymentsWithProgress() <-chan types.Progress {
	parts := 100_000
	buff := len(s.payments) + 1
	ch := make(chan types.Progress, buff)
	wg := sync.WaitGroup{}

	counter := 0
	for {
		beg := counter * parts
		end := (counter + 1) * parts
		if end > len(s.payments) {
			end = len(s.payments)
		}
		wg.Add(1)
		go func(ch chan types.Progress, data []*types.Payment) {
			defer wg.Done()
			progress := types.Progress{}

			for _, pay := range data {
				progress.Result += pay.Amount
			}
			progress.Part = 1
			ch <- progress
		}(ch, s.payments[beg:end])

		if end == len(s.payments) {
			break
		}
		counter++
	}
	wg.Wait()
	close(ch)
	return ch
}
