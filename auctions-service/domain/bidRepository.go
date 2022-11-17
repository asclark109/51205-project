package domain

type BidRepository interface {
	GetBid(bidId string) *Bid
	GetBidsByUserId(userId string) *[]*Bid
	GetBidsByItemId(itemId string) *[]*Bid
	SaveBid(bid *Bid)
	NextBidId() string
}
