package marketcap

import (
	"context"
	"errors"
	dtos2 "github.com/DaniilOr/marketcap/cmd/dtos"
	"github.com/DaniilOr/marketcap/pkg/updater"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math"
	"sort"
	"time"
)

var errNoDates = errors.New("array of dates is empty")
type Service struct{
	db *mongo.Database
	defaultStableCoins []string
	uniqueList []string

}
type timeSlice []time.Time

func CreateService(db *mongo.Database) *Service {
	var def = []string{"AMPL", "DGX", "DAI", "USDT", "USDC", "PAX",  "TUSD", "DAI", "USDK", "SAI", "EURS", "BITCNY", "GUSD", "SUSD", "USDS", "BGBP", "NUSD", "USNBT", "CONST", "BITUSD", "PESO", "EBASE", "HUSD",  "THKD", "WBTC"}
	return &Service{db: db, defaultStableCoins: def, uniqueList: updater.GetUniqueListData()}
}


func (s*Service) GetMarketCaps(ctx context.Context) (*[]dtos2.MarketCap, error){
	cursor, err := s.db.Collection("marketcap").Find(ctx, bson.M{})
	if err != nil{
		log.Println(err)
		return nil, err
	}
	var mcs []dtos2.MarketCap
	for cursor.Next(ctx){
		var mc dtos2.MarketCap
		err = cursor.Decode(&mc)
		if err != nil{
			log.Println(err)
			return nil, err
		}
		mcs = append(mcs, mc)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return &mcs, nil
}

func (s*Service) GetUniqueList(ctx context.Context) ([]string, error){
	cursor, err := s.db.Collection("prices").Find(ctx, bson.M{
		"$and": []bson.M{
			bson.M{"info.quoteAsset": "BTC"},
			bson.M{"info.status": "TRADING"},
		}})
	if err != nil{
		log.Println(err)
		return nil, err
	}
	var symbols []string
	for cursor.Next(ctx){
		var symbol dtos2.Price
		err = cursor.Decode(&symbol)
		if err != nil{
			log.Println(err)
			return nil, err
		}
		symbols = append(symbols, symbol.Symbol)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return symbols, nil
}

func (s*Service) getMarketByDate(ctx context.Context, date string, period int64)(*[]dtos2.MarketCap, error){
	findOptions := options.Find()
	findOptions.SetLimit(period + 1)
	findOptions.SetSort(map[string]int{"date": -1})
	cursor, err := s.db.Collection("marketcap").Find(ctx, bson.M{
		"date": bson.M{"$lte": date},
	}, findOptions)
	if err != nil{
		log.Println(err)
		return nil, err
	}
	var mcs []dtos2.MarketCap
	for cursor.Next(ctx){
		var mc dtos2.MarketCap
		err = cursor.Decode(&mc)
		if err != nil{
			log.Println(err)
			return nil, err
		}
		mcs = append(mcs, mc)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return &mcs, nil

}
func (s*Service) Recalculate(ctx context.Context, rebalancingPeriod int64, reconstitutionPeriod int64, startDate string, stableCoins []string, count int64, coins []string, reconstitute bool)(*dtos2.Response, error){
	if stableCoins == nil{
		stableCoins = s.defaultStableCoins
	}
	var periodDays int64 = rebalancingPeriod
	if reconstitute{
		periodDays = reconstitutionPeriod * rebalancingPeriod
	}
	var today string
	if len(startDate) == 0{
		t := time.Now()
		today = t.Format("2006-01-02")
	} else {
		today = startDate
	}
	res, err := s.getMarketByDate(ctx, today, periodDays)
	if err != nil{
		log.Println(err)
		return nil, err
	}
	topCaps := make(map[time.Time][]dtos2.PriceSymbolPair)
	var arrayOfDates timeSlice
	var used = make(map[string]bool)
	for _, i :=  range *res {
		for j := range i.Symbol{
				if !find(stableCoins, i.Symbol[j]) && find(s.uniqueList, i.Symbol[j]) {
					pair := dtos2.PriceSymbolPair{Symbol: i.Symbol[j], Cap: i.USD_marketcaps[i.Symbol[j]]}
					realDate, err := time.Parse("2006-01-02", i.Date)
					if err != nil{
						log.Println(err)
						return nil, err
					}
					topCaps[realDate] = append(topCaps[realDate], pair)
					if _, ok := used[i.Date]; !ok{
						arrayOfDates = append(arrayOfDates, realDate)
						used[i.Date] = true
					}
				}
		}
	}
	if len(arrayOfDates) == 0{
		return nil, errNoDates
	}
	firstDay := int64(len(arrayOfDates))-periodDays
	current := topCaps[arrayOfDates[firstDay]]
	var reversedArray timeSlice
	for i := range arrayOfDates {
		n := arrayOfDates[len(arrayOfDates)-1-i]
		reversedArray = append(reversedArray, n)
	}
	arrayOfDates = reversedArray
	for i := range arrayOfDates{
		if arrayOfDates[i].Format("2006-01-02") == today{
			continue
		}
		current = intersection(current, topCaps[arrayOfDates[i]])
	}
	var availableCoins []string
	for _, i := range current{
		availableCoins = append(availableCoins, i.Symbol)
	}
	if (!reconstitute) && (int(count) <= len(coins)){
		availableCoins = intersectArrays(availableCoins, coins)
	}
	// теперь надо для каждого коина собрать массв, где его marketcap-ы будут отсортированы по дате
	var coinsMarket = make(map[string][]float64)
	for _, i := range arrayOfDates{
		for _, j := range topCaps[i]{
			if i.Format("2006-01-02") == today{
				continue
			}
			if find(availableCoins, j.Symbol){
				coinsMarket[j.Symbol] = append(coinsMarket[j.Symbol], j.Cap)
			}
		}
	}
	var marketCaps = make(map[string]float64)
	for key, val := range coinsMarket{
		marketCaps[key] = wma(val)
	}
	filteredMarketCaps := transformerWrapper(marketCaps)
	filteredMarketCaps.Values = filteredMarketCaps.Values[int64(len(filteredMarketCaps.Values))-count:]
	filteredMarketCaps.Keys = filteredMarketCaps.Keys[int64(len(filteredMarketCaps.Keys))-count:]
	println(len(filteredMarketCaps.Keys))
	marketCaps = make(map[string]float64)
	for i, key := range filteredMarketCaps.Keys{
		marketCaps[key] = filteredMarketCaps.Values[i]
	}
	functionResult := averageWeight(marketCaps)
	resp := s.wrapResponse(functionResult.Keys, functionResult.Values)
	return  &resp, nil
}


func find(arr []string, target string) bool{
	for _, i:=range arr{
		if i== target{
			return true
		}
	}
	return false
}

func wma(ps []float64) float64{
	denom := (1 + float64(len(ps))) / 2 * float64(len(ps))
	sum := 0.0
	for i, p := range ps{
		sum += float64((i+1)) * p
	}
	return sum / denom
}

func sqrSum(ps []float64) float64{
	sum := 0.0
	for _, i := range ps{
		sum += math.Sqrt(i)
	}
	return sum
}

func averageWeight(ps map[string]float64) dtos2.RebalancingResult {
	var wmas = make([]float64, 0)
	for _, val := range ps{
		wmas = append(wmas, val)
	}
	sqrtsumma := sqrSum(wmas)
	for key, val := range ps{
		ps[key] = math.Sqrt(val) / sqrtsumma
	}
	return transformerWrapper(ps)
}

func transformerWrapper(ps map[string]float64) dtos2.RebalancingResult {
	vals := []float64{}
	keys := []string{}
	n := map[float64][]string{}
	var a []float64
	for key, val := range ps {
		n[val] = append(n[val], key)
	}
	for k := range n {
		a = append(a, k)
	}
	sort.Float64s(a)
	for _, k := range a {
		for _, s := range n[k] {
			vals = append(vals, k)
			keys = append(keys, s)
		}
	}
	return dtos2.RebalancingResult{Keys: keys, Values: vals}
}

func intersection(a, b []dtos2.PriceSymbolPair) (c []dtos2.PriceSymbolPair) {
		set1 := make(map[string]bool)
		set2 := make(map[string]bool)
		used := make(map[string]bool)
		var allPairs = make([]dtos2.PriceSymbolPair, 0)
		for _, item := range a {
			set1[item.Symbol] = true
			allPairs = append(allPairs, item)

		}
		for _, item := range b {
			set2[item.Symbol] = true
			allPairs = append(allPairs, item)

		}

	for _, item := range allPairs{
			_, ok1 := set1[item.Symbol]
			_, ok2 := set2[item.Symbol]
			if ok1 && ok2 && !used[item.Symbol]{
				c = append(c, item)
				used[item.Symbol] = true
			}
		}
		return c
}

func (p timeSlice) Len() int {
	return len(p)
}

func (p timeSlice) Less(i, j int) bool {
	return p[i].Unix()  < p[j].Unix()
}

func (p timeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func intersectArrays(a, b []string) (c []string) {
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)
	used := make(map[string]bool)
	var allPairs = make([]string, 0)
	for _, item := range a {
		set1[item] = true
		allPairs = append(allPairs, item)

	}
	for _, item := range b {
		set2[item] = true
		allPairs = append(allPairs, item)
	}
	for _, item := range allPairs{
		_, ok1 := set1[item]
		_, ok2 := set2[item]
		if ok1 && ok2 && !used[item]{
			c = append(c, item)
			used[item] = true
		}
	}
	return c
}

func (s*Service) wrapResponse(keys []string, vals []float64) dtos2.Response {
	var resp dtos2.Response
	resp.RecalculatedWeights = make(map[string]float64, 0)
	for i, _ := range keys{
		resp.RecalculatedWeights[keys[i]] = vals[i]
	}
	return resp
}

func (s*Service) updateAll(){
	s.uniqueList = updater.GetUniqueListData()
	MCS := updater.GetMarketcapData()
	count, err := s.db.Collection("marketcap").CountDocuments(context.Background(), bson.M{"date": MCS.Date})
	if err != nil{
		log.Println(err)
		return
	}
	log.Println(count)
	if count == 0{
		obj, err := s.db.Collection("marketcap").InsertOne(context.Background(), MCS)
		if err != nil{
			log.Println(err)
		} else {
			log.Println(obj.InsertedID)
		}
	} else {
		id, err := s.db.Collection("marketcap").UpdateOne(context.Background(),
			bson.M{"date": bson.M{"$eq": MCS.Date }},
			bson.D{
				{"$set", bson.D{{"symbols", MCS.Symbol}}},
				{"$set", bson.D{{"USD_marketcaps", MCS.USD_marketcaps}}},
			},)
		if err != nil{
			log.Println(err)
		} else {
			log.Println(id)
		}
	}
}

func (s *Service) StartScrapping() error{
	c := cron.New()
	_, err := c.AddFunc("@every 1h", s.updateAll)
	if err != nil{
		log.Println(err)
		return err
	}
	c.Start()
	return nil
}