package main

import (
  "time"
  "math/rand"
  "math"
  "log"
	"net/http"
  "context"
	"encoding/json"
	"errors"
  "flag"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type BiddingService interface {
  ValidateAdPlacementId(string) bool
  ShouldBid(string) bool
  SelectAd(string) string
  ComputeBid(string) float64
}

type biddingService struct {}

type biddingRequest struct {
	AdPlacementId string `json:"AdPlacementId"`
}

type biddingResponse struct {
	AdId   string `json:"AdId"`
	BidPrice float64 `json:"BidPrice"`
}

func (biddingService) ValidateAdPlacementId(adPlacementId string) bool {
  // We don't have an algorithm to make this determination yet so
  // for the moment we just accept every id as valid
  return true
}

func (bs biddingService) ShouldBid(adPlacementId string) bool {
  // We will randomly choose whether to bid or not.
  if rand.Intn(2) == 1 && bs.ValidateAdPlacementId(adPlacementId) {
    return true
  }
  return false
}

func (biddingService) ComputeBid(adId string) float64 {
  // We decide bid price in a random fashion for the moment.
  // There could be various ways to do this like having a list of prices to choose
  // from.
  return math.Round(rand.Float64()*100)/100
}

func (biddingService) SelectAd(adPlacementId string) string {
  // Ideally we would do some sorta DB query here to obtain ad's suitable for
  // the placement slot. For the time we are going to randomly choose from a
  // pre-defined list of ad ids.
  ad_ids := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
  return ad_ids[rand.Intn(10)]
}

func makeBidEndpoint(bvc BiddingService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(biddingRequest)
		if bvc.ShouldBid(req.AdPlacementId) {
      adId := bvc.SelectAd(req.AdPlacementId)
      bid_price := bvc.ComputeBid(adId)
      return biddingResponse{adId, bid_price}, nil
    } else {
      return nil, nil
    }
	}
}

func main() {
  var (
		port = flag.String("port", "8080", "HTTP listen at port")
	)
  flag.Parse()

  rand.Seed(time.Now().UnixNano())
  bvc := biddingService{}

  bidHandler := httptransport.NewServer(
    makeBidEndpoint(bvc),
    decodeBiddingRequest,
		encodeResponse,
  )

  http.Handle("/placebid", bidHandler)
  log.Println("Starting bidding service at port " + *port)
	log.Fatal(http.ListenAndServe(":" + *port, nil))
}

func decodeBiddingRequest(_ context.Context, r *http.Request) (interface{}, error) {
  if r.Method != "POST" {
    return nil, errors.New("Only POST requests allowed.")
  }
	var request biddingRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(c context.Context, w http.ResponseWriter, response interface{}) error {
  if response == nil {
    w.WriteHeader(http.StatusNoContent)
    w.Header().Set("Content-Type", "text/plain")
    return nil
  } else {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    return json.NewEncoder(w).Encode(response)
  }
}
