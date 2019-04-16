package main

import (
  "time"
  "log"
	"net/http"
  "context"
	"encoding/json"
	"errors"
  "bytes"
  "fmt"
  "io/ioutil"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type AuctionService interface {
  ValidateAdPlacementId(string) bool
  GetBiddingServices(string) []biddingService
  ConductAuction(string) auctionResponse
  MakeBidRequest(biddingService, string, chan auctionResponse) error
}

type auctionService struct {}

type biddingService struct {
  host string
  port string
}

type auctionRequest struct {
	AdPlacementId string `json:"AdPlacementId"`
}

type auctionResponse struct {
	AdId   string `json:"AdId"`
	BidPrice float64 `json:"BidPrice"`
}

func (auctionService) ValidateAdPlacementId(adPlacementId string) bool {
  // We don't have an algorithm to make this determination yet so
  // for the moment we just accept every id as valid
  return true
}

func (auctionService) GetBiddingServices(adPlacementId string) []biddingService {
  // Ideally we would do some sorta DB query here to obtain compatible bidding
  // services but for now lets just use some static addresses.
  bidding_services := []biddingService{
    {
      "localhost",
      "8080",
    },
    {
      "localhost",
      "8080",
    },
  }
  return bidding_services
}

func (auctionService) MakeBidRequest(bs biddingService, adPlacementId string, bids chan auctionResponse) error {
  url := fmt.Sprintf("http://%s:%s/placebid", bs.host, bs.port)
  jsonValue, _ := json.Marshal(auctionRequest{adPlacementId})
  timeout := time.Duration(200 * time.Millisecond)
  client := http.Client{
    Timeout: timeout,
  }
  httpResponse, err := client.Post(url, "application/json", bytes.NewBuffer(jsonValue))

  if err != nil {
    fmt.Println(err)
    return err
  }
  defer httpResponse.Body.Close()
  if httpResponse.StatusCode == http.StatusOK {
    var auction_response auctionResponse
    body, _ := ioutil.ReadAll(httpResponse.Body)
    if err := json.Unmarshal(body, &auction_response); err != nil {
      fmt.Println(err)
      return err
    }
    bids <- auction_response
  } else {
    bids <- auctionResponse{}
  }
  return nil
}

func (as auctionService) ConductAuction(adPlacementId string) auctionResponse {
  bidding_services := as.GetBiddingServices(adPlacementId)

  bids := make(chan auctionResponse)

  for _, bs := range bidding_services {
    go as.MakeBidRequest(bs, adPlacementId, bids)
  }

  max_bid_resp := auctionResponse{}
  count := 0
  for count != len(bidding_services) {
    bid := <-bids
    count++
    if max_bid_resp.AdId != "" {
      if bid.AdId != "" && bid.BidPrice > max_bid_resp.BidPrice {
        max_bid_resp = bid
      }
    } else {
      max_bid_resp = bid
    }
  }

  if max_bid_resp.BidPrice >= 0 {
    return max_bid_resp
  } else {
    return auctionResponse{}
  }
}

func makeAuctionEndpoint(as AuctionService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(auctionRequest)
		if as.ValidateAdPlacementId(req.AdPlacementId) {
      res := as.ConductAuction(req.AdPlacementId)
      if res.AdId == "" {
        return nil, nil
      }
      return res, nil
    } else {
      return nil, nil
    }
	}
}

func main() {
  as := auctionService{}

  auctionHandler := httptransport.NewServer(
    makeAuctionEndpoint(as),
    decodeAuctionRequest,
		encodeResponse,
  )

  http.Handle("/auction", auctionHandler)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func decodeAuctionRequest(_ context.Context, r *http.Request) (interface{}, error) {
  if r.Method != "POST" {
    return nil, errors.New("Only POST requests allowed.")
  }
	var request auctionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(c context.Context, w http.ResponseWriter, response interface{}) error {
  if response == nil {
    c = httptransport.SetResponseHeader("status", "204")(c, w)
    c = httptransport.SetContentType("text/plain")(c, w)
    return nil
  } else {
    c = httptransport.SetContentType("application/json")(c, w)
    c = httptransport.SetResponseHeader("status", "200")(c, w)
    return json.NewEncoder(w).Encode(response)
  }
}
