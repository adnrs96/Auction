package main

import (
  "net/http"
  "net/http/httptest"

  "testing"
  "strings"
  "fmt"
	"path/filepath"
	"runtime"
	"reflect"
  "encoding/json"
)

func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func TestMakeBidRequest(t *testing.T) {
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		equals(t, req.URL.String(), "/placebid")
    response := AuctionResponse{"1", float64(0.22)}
		json.NewEncoder(w).Encode(response)
	}))

	defer server.Close()

  bids := make(chan AuctionResponse)
  bs := BiddingService{strings.Split(server.URL, "/")[2]}
  adPlacementId := "12345"
  as := auctionService{}

  go as.MakeBidRequest(bs, adPlacementId, bids)

  bid := <- bids

  equals(t, bid.AdId, "1")
  equals(t, bid.BidPrice, float64(0.22))
}
