package app

import (
	"encoding/json"
	"github.com/DaniilOr/marketcap/cmd/dtos"
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
	s.router.Post("/recalculate_weights", s.recalculate)
	//s.router.Get("/test", s.test)
	err := s.marketcapSvc.StartScrapping()
	if err != nil{
		log.Println(err)
		return err
	}
	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func (s*Server) recalculate(writer http.ResponseWriter, request *http.Request){
	decoder := json.NewDecoder(request.Body)
	var requestParameters dtos.RequestJSON
	err := decoder.Decode(&requestParameters)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("%v",requestParameters)
	res, err := s.marketcapSvc.Recalculate(request.Context(), requestParameters.RebalancingPeriod, requestParameters.ReconstitutionPeriod, requestParameters.StartDate, requestParameters.StableCoins, requestParameters.Count, requestParameters.Coins, requestParameters.Reconstitution )
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