package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type LogPlay struct {
	UID        string
	Name       string
	Bet        int
	MyCard     string
	DealerCard string
	IsDealer   bool
	ChipGain   int
	DealerUID  string
}

type LogDealerPlay struct {
	BetRate int
	Value   int
}

type Payment struct {
	UID     string
	Request string
	Time    string
	IsPass  bool
}

func SaveLogDealer(p LogDealerPlay) {
	bytesRepresentation, _ := json.Marshal(p)
	http.Post("http://10.128.0.2:3111/log_dealer_pay", "application/json", bytes.NewBuffer(bytesRepresentation))
}

func SaveLogPlay(p LogPlay) {
	bytesRepresentation, _ := json.Marshal(p)
	http.Post("http://10.128.0.2:3111/log_pay", "application/json", bytes.NewBuffer(bytesRepresentation))
}

func RequestPayment(ctx context.Context, uid string, num string, times string) (bool, error) {
	go func() {
		select {
		case <-time.After(60 * time.Second):
			fmt.Println("overslept")
		case <-ctx.Done():
			fmt.Println(ctx.Err()) // prints "context deadline exceeded"
		}
	}()
	tr := &http.Transport{
		ResponseHeaderTimeout: 60 * time.Second,
	}
	client := &http.Client{
		Timeout:   50 * time.Second,
		Transport: tr,
	}
	url := "http://10.128.0.2:3111/payment/" + uid + "/" + num + "/" + times
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err

	}

	resp, err := client.Do(req.WithContext(ctx))
	// resp, err := http.Get("http://10.128.0.2:3000/payment/" + uid + "/" + num + "/" + time)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return true, nil
	} else if resp.StatusCode == 400 {
		return false, nil
	} else {
		return false, errors.New(fmt.Sprintf("Status %v", resp.StatusCode))
	}

}

func SaveBank(email string, no string, name string) error {
	resp, err := http.Get("http://10.128.0.2:3111/bank/" + email + "/" + no + "/" + name)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		return nil
	} else {
		return errors.New("status not 200")
	}

}

func SaveTax(value int) {
	http.Get("http://10.128.0.2:3111/tax/" + strconv.Itoa(value))
}
