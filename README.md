# Auction!

Hey there! So this is my very first experience with writing `go` code ever. Inside this project is a very basic auction service designed using with keeping micro service design in mind. We are designing these two services to conduct an auction for a given advertisement slot. I guess this also makes this project my first ever i-tty bitty experience with AdTech world. Code consists of two so called micro services. First is the `bidding` and the other one is `auctioneer`. As their names might suggest, `bidding` service is used to place bids for a given Ad Slot and the `auctioneer` service is responsible for conducting this entire auction by accepting Ad Slot's (items) to be auctioned and calls a suitable set of bidding services for placing bids for the slot. Finally the `auctioneer` service acts as to whom the item (Ad Slot) in question is to be finally auctioned.


# Instructions to get things running...

To be honest I don't feel there it should be super tedious to get things running. After all I have included a docker-compose file along with and took the time to create service image DockerFile's. I believe service can be started using the following command:
```
docker-compose up
```
Don't forget to use `sudo` if you need it.

Actually in general one should first build the service images using
```
docker-compose build
```
but thanks to docker you can avoid that since `docker-compose up` builds these images automatically the first time since they probably won't exist for you the first time. Be careful that after building these images if you ever end up modifying the code and want to run it, make sure you first build it and then run the images using `docker-compose up`. Alternately as a short cut you can say
```
docker-compose up --build
```


## Testing

I guess I wasn't able to do it all in the allocated time period. You can run the test I wrote using
```
go test
```
from inside `Auction/cmd/auctioneer/`. Ideally I would have wanted to make docker run tests but for the timing this would have to do.
