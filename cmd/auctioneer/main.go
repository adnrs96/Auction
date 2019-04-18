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
  "flag"
  "os"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

var (
  port = flag.String("port", "8080", "HTTP listen at port")
)

type AuctionService interface {
  ValidateAdPlacementId(string) bool
  GetBiddingServices(string) []biddingService
  ConductAuction(string) auctionResponse
  MakeBidRequest(biddingService, string, chan auctionResponse)
}

type auctionService struct {}

type biddingService struct {
  Addr string
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
  // services but for now lets just accept bidding service host from CLI.
  biddingAddress := os.Getenv("BIDDING_SERVICE_URL")
  if biddingAddress == "" {
    biddingAddress = "biddingservice:8080"
  }
  bidding_services := []biddingService{
    {biddingAddress},
  }
  return bidding_services
}

func (auctionService) MakeBidRequest(bs biddingService, adPlacementId string, bids chan auctionResponse) {
  url := fmt.Sprintf("http://%s/placebid", bs.Addr)
  jsonValue, _ := json.Marshal(auctionRequest{adPlacementId})
  timeout := time.Duration(200 * time.Millisecond)
  client := http.Client{
    Timeout: timeout,
  }
  log.Printf("Making a bid request to %s for a bid for AdPlacementId %s", url, adPlacementId)
  httpResponse, err := client.Post(url, "application/json", bytes.NewBuffer(jsonValue))

  if err != nil {
    bids <- auctionResponse{}
    return
  }
  defer httpResponse.Body.Close()
  if httpResponse.StatusCode == http.StatusOK {
    var auction_response auctionResponse
    body, _ := ioutil.ReadAll(httpResponse.Body)
    if err := json.Unmarshal(body, &auction_response); err != nil {
      bids <- auctionResponse{}
    } else {
      bids <- auction_response
    }
  } else {
    bids <- auctionResponse{}
  }
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
  flag.Parse()
  as := auctionService{}

  auctionHandler := httptransport.NewServer(
    makeAuctionEndpoint(as),
    decodeAuctionRequest,
		encodeResponse,
  )

  http.Handle("/auction", auctionHandler)
  log.Println("Starting bidding service at port " + *port)
	log.Fatal(http.ListenAndServe(":" + *port, nil))
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
