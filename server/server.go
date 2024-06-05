package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type UsdBrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type EconomiaApiRespose struct {
	UsdBrl UsdBrl `json:"USDBRL"`
}

type ApiError struct {
	Message string `json:"message"`
}

type ApiResponse struct {
  Bid string `json:"bid"`
}

func main() {
  http.HandleFunc("/cotacao", exchangeRateHandler)
	http.ListenAndServe(":8080", nil)
}

func exchangeRateHandler(w http.ResponseWriter, r *http.Request) {
	exchangeRate, err := getExchangeRate()
	if err != nil {
		fmt.Println(err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(ApiError{
			Message: "Failed to get exchange rate information",
		})
		return
	}

  err = persistExchangeRate(exchangeRate)
  if err != nil {
    fmt.Println(err)

    w.WriteHeader(http.StatusInternalServerError)
    
    json.NewEncoder(w).Encode(ApiError{
      Message: "Failed to persist the exchange rate information",
    })
    return
  }

  response := ApiResponse{
    Bid: exchangeRate.UsdBrl.Bid,
  }

  w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func getExchangeRate() (*EconomiaApiRespose, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return &EconomiaApiRespose{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &EconomiaApiRespose{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &EconomiaApiRespose{}, err
	}

	var parsedResponse EconomiaApiRespose
	err = json.Unmarshal(body, &parsedResponse)
	if err != nil {
		return &EconomiaApiRespose{}, err
	}

	return &parsedResponse, nil
}

func persistExchangeRate(rate *EconomiaApiRespose) error {
  ctx := context.Background()
  ctx, cancel := context.WithTimeout(ctx, time.Millisecond * 20)
  defer cancel()

  db, err := sql.Open("sqlite3", "./exchange.db") 
  if err != nil {
    return err
  }

  defer db.Close()

  createTableStmt, err := db.PrepareContext(ctx, `
  create table if not exists exchange_rate (id text primary key, bid text);
  `)
  if err != nil {
    return err
  }

  defer createTableStmt.Close()

  createTableStmt.ExecContext(ctx)

  stmt, err := db.Prepare("insert into exchange_rate(id, bid) values(?, ?)")
  if err != nil {
    return err
  }

  defer stmt.Close()

  _, err = stmt.ExecContext(ctx, uuid.New().String(), rate.UsdBrl.Bid)
  if err != nil {
    return err
  }

  return nil
}
