package main

import (
	"account"
	"data"
	"dbfunc"
	"fmt"
	"logger"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	logger.Now = time.Now().Format(time.RFC822)
	logger := logger.GetLogger("[Let's get Rich]")
	logger.Println("Let's Get Start!")

	var myAccounts *account.MyBalance
	var myLimitOrders *account.MyLimitOrders
	db := dbfunc.GetDbConn("BTC")

	// get Account Info every 10 seconds
	go func() {
		myAccounts = account.GetBalance()
		myLimitOrders = myAccounts.GetLimitOrders("BTC")
		time.Sleep(time.Duration(10) * time.Second)
	}()
	// get BTC Trade data every 10 minutes.
	go func() {
		for {
			data.GetCoinTradeData("BTC", db)
			time.Sleep(time.Duration(10) * time.Minute)
		}
	}()

	time.Sleep(time.Duration(200) * time.Minute)

	go func() {
		for {
			ctp := dbfunc.Select(db, "BTC", 5)
			tangent := float64((ctp[0].Bolband-ctp[1].Bolband)/ctp[0].Bolband) + 0.005
			ro := data.GetRecentOrder("BTC")
			currentValue, _ := strconv.ParseUint(ro.Ask.Price, 10, 64)
			if currentValue < (ctp[0].Bolband-5*uint64(ctp[0].Bolbandsd)/2) && tangent > 0 {
				weight := (tangent * 100) * 0.5
				available, _ := strconv.ParseFloat(myAccounts.Krw.Available, 64)
				qty := available * weight / float64(currentValue)
				buyID := myAccounts.BuyCoin("BTC", currentValue, qty)

				time.Sleep(time.Duration(15) * time.Minute)
				var i int
				for _, limitOrder := range myLimitOrders.LimitOrders {
					if limitOrder.OrderId == buyID {
						myAccounts.CancelOrder(buyID, currentValue, qty, "bid")
					} else {
						i++
					}
				}
				if i == len(myLimitOrders.LimitOrders) {
					ro = data.GetRecentOrder("BTC")
					currentValue, _ := strconv.ParseUint(ro.Bid.Price, 10, 64)
					myAccounts.SellCoin("BTC", currentValue, qty)
				}

			}
			time.Sleep(time.Duration(5) * time.Second)
		}
	}()

	time.Sleep(time.Duration(10) * time.Second)
	go func() {
		currentTime := time.Now().Unix()
		for _, limitOrder := range myLimitOrders.LimitOrders {
			timestamp, _ := strconv.ParseInt(limitOrder.Timestamp, 10, 64)
			if timestamp < currentTime-3600 {
				price, _ := strconv.ParseUint(limitOrder.Price, 10, 64)
				qty, _ := strconv.ParseFloat(limitOrder.Qty, 64)
				myAccounts.CancelOrder(limitOrder.OrderId, price, qty, limitOrder.Type)
			}
		}
	}()

	fmt.Scanln()
}
