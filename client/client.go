package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ApiResponse struct {
  Bid string `json:"bid"`
}

const outputFileName string = "cotacao.txt"

func main() {
  exchangeRage, err := getExchangeRate()
  if err != nil {
    panic(err)
  }

  err = persistExchangeRate(exchangeRage)
  if err != nil {
    panic(err)
  }

  fmt.Printf("Cotação salva em '%s'", outputFileName)
}

func getExchangeRate() (*ApiResponse, error) {
  ctx := context.Background()
  ctx, cancel := context.WithTimeout(ctx, time.Millisecond * 300)
  defer cancel()

  req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
  if err != nil {
    return &ApiResponse{}, err 
  }

  res, err := http.DefaultClient.Do(req)
  if err != nil {
    return &ApiResponse{}, err
  }

  defer res.Body.Close()

  body, err := io.ReadAll(res.Body)
  if err != nil {
    return &ApiResponse{}, err
  }

  var parsedResponse ApiResponse
  err = json.Unmarshal(body, &parsedResponse)
  if err != nil {
    return &ApiResponse{}, err
  }

  return &parsedResponse, nil
}

func persistExchangeRate(exchangeRate *ApiResponse) error {
  file, err := os.Create(outputFileName)
  if err != nil {
    return err
  }

  defer file.Close()
  fileContent := fmt.Sprintf("Dólar: %s", exchangeRate.Bid)

  _, err = file.WriteString(fileContent)
  return err
}
