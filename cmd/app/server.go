package app

import (
	"encoding/json"
	marketcap2 "github.com/DaniilOr/marketcap/pkg/marketcap"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)


type Server struct{
	marketcapSvc *marketcap2.Service
	router chi.Router

}

func NewServer(mc *marketcap2.Service, router chi.Router) *Server {
	return &Server{marketcapSvc: mc, router: router}
}


func (s *Server) Init() error {
	s.router.Use(middleware.Logger)
	s.router.Get("/recalculate_weights", s.recalculate)
	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func (s*Server) recalculate(writer http.ResponseWriter, request *http.Request){
	var stableCoins = []string{
		"USDT", "NUSD", "THKD", "PESO", "USDC", "DAI", "BUSD", "TUSD", "HUSD", "PAX", "USDK", "EURS", "GUSD", "SUSD", "USDS", "WBTC",
	}
	var coins = []string{
		"ETH", "BNB", "DOGE", "XRP", "ADA", "DOT", "BCH", "UNI", "LTC", "LINK", "XLM", "VET", "SOL", "THETA", "FIL", "ETC", "TRX", "EOS", "XMR", "MATIC", "NEO", "AAVE", "LUNA", "CAKE", "FTT", "ATOM", "XTZ", "MKR", "AVAX", "ALGO",	}
	res, err := s.marketcapSvc.Recalculate(request.Context(), 14, 2, "2021-05-28", stableCoins, 30, coins, false )
	if err != nil{
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
	body, err := json.Marshal(res)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(body)
	if err != nil {
		log.Println(err)
	}
}