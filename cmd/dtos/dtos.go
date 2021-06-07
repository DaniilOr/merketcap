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

type RequestJSON struct{
	StableCoins []string `json:"stable_coins"`
	Coins []string `json:"coins"`
	RebalancingPeriod int64 `json:"rebalancing_period"`
	ReconstitutionPeriod int64 `json:"reconstitution_period"`
	StartDate string `json:"start_date"`
	Count int64 `json:"count"`
	Reconstitution bool `json:"reconstitution"`
}

type Response struct {
	RecalculatedWeights map[string]float64 `json:"recalculated_weights"`
}