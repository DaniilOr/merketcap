package updater

import (
	"context"
	"encoding/json"
	"github.com/DaniilOr/marketcap/cmd/dtos"
	"github.com/adshao/go-binance/v2"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)


type ParsedData struct{
	MC float64 `json:"quote.USD.market_cap"`
	Symbol string `json:"symbol"`
}
type InitialResp struct{
	Status map[string]interface{} `json:"status"`
	Data interface{} `json:"data"`
}
func GetMarketcapData() (*dtos.MarketCap){
	url := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	req.Header.Set("Accepts", "application/json")
	req.Header.Set("X-CMC_PRO_API_KEY", "3392a31e-0852-4417-8366-1e210ca806ba")
	req.URL.Query().Add("start", "1")
	req.URL.Query().Add( "limit", "400")
	req.URL.Query().Add( "convert", "USD")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil{
		log.Println(err)
		return nil
	}
	var t InitialResp
	data, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &t); err != nil {
		log.Fatalf("Parse response failed, reason: %v \n", err)
	}
	var parsedMC dtos.MarketCap
	parsedMC.USD_marketcaps = make(map[string]float64)
	parsedMC.Symbol = make([]string, 0)

	for _, i := range t.Data.([]interface{}){
		representation := i.(map[string]interface{})
		quote := representation["quote"].(map[string]interface{})
		USD := quote["USD"].(map[string]interface{})
		parsedMC.USD_marketcaps[representation["symbol"].(string)] = USD["market_cap"].(float64)
		parsedMC.Symbol = append(parsedMC.Symbol, representation["symbol"].(string))
	}
	parsedMC.Date = time.Now().Format("2006-01-02")
	return &parsedMC
}
func GetUniqueListData() []string{
	var (
		apiKey = ""
		secretKey = ""
	)
	client := binance.NewClient(apiKey, secretKey)
	info, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil{
		log.Println(err)
		return nil
	}
	var uniqueList []string
	for _, s := range info.Symbols{
		if (s.Status == "TRADING") && (s.QuoteAsset == "BTC"){
			uniqueList = append(uniqueList, s.BaseAsset)
		}
	}
	uniqueList = append(uniqueList, "BTC")
	return uniqueList
}