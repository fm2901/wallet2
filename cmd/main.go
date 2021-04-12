package main

import (
	"fmt"
	"github.com/fm2901/wallet/pkg/wallet"
)


func main() {
	svc := &wallet.Service{}
	account, err := svc.RegisterAccount("+992926409000")

	err = svc.Deposit(account.ID, 100)
	

	payment, err := svc.Pay(account.ID, 10, "car")

	err = svc.Reject(payment.ID)
	
	fmt.Println(err)
}