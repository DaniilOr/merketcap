package dtos

type Price struct {
	Price string `bson:"price"`
	Symbol string `bson:"symbol" json:"symbol"`
}

type MarketCap struct {
	USD_marketcaps map[string]float64 `bson:"USD_marketcaps"`
	Id int64 `bson:"id"`
	Date string `bson:"date"`
	Symbol []string `bson:"symbols"`
}

type PriceSymbolPair struct {
	Symbol string `json:"symbol"`
	Cap float64 `json:"cap"`
}

type RebalancingResult struct {
	Keys []string `json:"keys"`
	Values []float64 `json:"values"`
}