FROM golang:1.11 as builder

LABEL maintainer="Aditya Bansal <adi.bansal241996@gmail.com>"

WORKDIR /go/src/github.com/adnrs96/Auction/cmd/auctioneer

COPY . .

RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/auctioneerservice .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /go/bin/auctioneerservice .

EXPOSE 8080

CMD ["./auctioneerservice"]
