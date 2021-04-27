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

//ErrPhoneRegistered - телефон уже регитрирован
var ErrPhoneRegistered = errors.New("phone already registred")

//ErrAmountMustBePositive - счёт не может быть пустым
var ErrAmountMustBePositive = errors.New("amount must be greater than zero")

//ErrAccountNotFound - пользователь не найден
var ErrAccountNotFound = errors.New("account not found")

//ErrNotEnoughtBalance - на счете недостаточно средств
var ErrNotEnoughtBalance = errors.New("account not enough balance")

//ErrPaymentNotFound - платеж не найден
var ErrPaymentNotFound = errors.New("payment not found")

// ErrFavoriteNotFound - Избранное не найдено
var ErrFavoriteNotFound = errors.New("favorite not found")

// Service представляет информацию о пользователе.
type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
	historys      []*types.Payment
}

// RegisterAccount - метод для регистрация нового прользователя.
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

//Pay метод для регистрации платижа
func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
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
		return nil, ErrNotEnoughtBalance
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

// FindAccountByID ищем пользователя по ID
func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
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

	return account, nil
}

// FindPaymentByID ищем платёж по ID
func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	var payment *types.Payment

	for _, pay := range s.payments {
		if pay.ID == paymentID {
			payment = pay
		}
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	return payment, nil
}

// FindFavoriteByID ищем платёж по ID в Избранное
func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}
	return nil, ErrFavoriteNotFound
}

//Deposit method
func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount < 0 {
		return ErrAmountMustBePositive
	}

	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return err
	}

	account.Balance += amount
	return nil
}

// Reject метод для отмены покупок
func (s *Service) Reject(paymentID string) error {
	pay, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}

	acc, err := s.FindAccountByID(pay.AccountID)
	if err != nil {
		return err
	}

	pay.Status = types.PaymentStatusFail
	acc.Balance += pay.Amount

	return nil
}

// Repeat позволāет по идентификатору повторитþ платёж
func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	pay, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	payment, err := s.Pay(pay.AccountID, pay.Amount, pay.Category)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// FavoritePayment добавления новых Избранных
func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	favoriteID := uuid.New().String()
	newFavorite := &types.Favorite{
		ID:        favoriteID,
		AccountID: payment.AccountID,
		Name:      name,
		Amount:    payment.Amount,
		Category:  payment.Category,
	}

	s.favorites = append(s.favorites, newFavorite)
	return newFavorite, nil
}

//PayFromFavorite для совершения платежа в Избранное
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

//ExportToFile экспортирует аккаунт в файл
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

//ImportFromFile импортировать с файла
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

//Export экспортировать
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

//Import импортирует данные из файла
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

//ExportAccountHistory вытаскивает платежи конкретного аккаунта
func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {
	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return nil, err
	}

	history := []types.Payment{}
	for _, payment := range s.payments {
		if payment.AccountID == account.ID {
			history = append(history, types.Payment{
				ID:        payment.ID,
				AccountID: payment.AccountID,
				Amount:    payment.Amount,
				Category:  payment.Category,
				Status:    payment.Status,
			})
		}
	}
	return history, nil
}

//HistoryToFile создаёт информацию о платеже в строки для файла
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

//HistoryToFiles помешает по record данные платежей в файл
func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {
	if len(payments) < 1 {
		return nil
	}
	if len(payments) <= records {
		HistoryToFile(payments, dir+"/payments.dump")
	} else {
		counter := 1
		fIndex := 0
		lIndex := records
		for {
			HistoryToFile(payments[fIndex:lIndex], dir+"/payments"+fmt.Sprint(counter)+".dump")
			fIndex += records
			lIndex += records
			if lIndex >= len(payments) {
				if counter*records < len(payments) {
					lIndex = len(payments) - counter*records
					HistoryToFile(payments[:lIndex], dir+"/payments"+fmt.Sprint(counter+1)+".dump")
				}
				break
			}
			counter++
		}
	}
	return nil
}

//SumPayments суммирует по горутине
func (s *Service) SumPayments(gorutines int) types.Money {
	wg := sync.WaitGroup{}
	wg.Add(gorutines)
	mu := sync.Mutex{}
	sum := types.Money(0)
	if gorutines < 2 {
		go func() {
			defer wg.Done()
			val := types.Money(0)
			for _, pay := range s.payments {
				val += pay.Amount
			}
			mu.Lock()
			sum += val
			defer mu.Unlock()

		}()

	}
	if gorutines > 1 {
		for i := 0; i < gorutines; i++ {
			go func() {
				defer wg.Done()
				val := types.Money(0)
				for _, pay := range s.payments {
					val += pay.Amount
				}
				mu.Lock()
				sum += val / 10
				defer mu.Unlock()

			}()
		}
	}
	wg.Wait()
	return sum
}

//FilterPayments фильтрует данные по accountID через gorutines
func (s *Service) FilterPayments(accountID int64, gorutines int) ([]types.Payment, error) {
	payments, err := s.ExportAccountHistory(accountID)
	if err != nil {
		return nil, err
	}
	payment := []types.Payment{}
	wg := sync.WaitGroup{}
	wg.Add(gorutines)
	mu := sync.Mutex{}
	if gorutines < 2 {
		go func() {
			defer wg.Done()
			for _, pay := range payments {
				pay.Amount = s.SumPayments(gorutines)
				payment = append(payment, pay)
			}
			mu.Lock()
			defer mu.Unlock()
		}()
	}
	if gorutines > 1 {
		for i := 0; i < gorutines*2; i++ {
			go func() {
				defer wg.Done()
				for _, pay := range payments {
					pay.Amount = s.SumPayments(gorutines)
					payment = append(payment, pay)
				}
				mu.Lock()
				defer mu.Unlock()
			}()
		}
	}
	return payments, nil
}

func (s *Service) FilterPaymentsByFn(filter func(payment types.Payment) bool, goroutines int) ([]types.Payment, error) {
	payments := []types.Payment{}

	if goroutines == 0 || goroutines == 1 {
		for _, payment := range s.payments {
			if filter(*payment) {
				payments = append(payments, *payment)
			}
		}
		return payments, nil
	}

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
					if filter(*payment) {
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
					if filter(*payment) {
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
