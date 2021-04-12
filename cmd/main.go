package main

import (
	"fmt"

	"github.com/fm2901/wallet/pkg/wallet")


func main() {
	svc := &wallet.Service{}
	account, err := svc.RegisterAccount("+992926409000")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(account)

	err = svc.Deposit(account.ID, 10)
	if err != nil {
		switch err {
		case wallet.ErrAmountmustBePositive:
			fmt.Println("Сумма должна быть позитивной")
		case wallet.ErrAccountnotFound:
			fmt.Println("Аккаунт пользователя не найден")
		}
		return
	}

	fmt.Println(account.Balance)
}