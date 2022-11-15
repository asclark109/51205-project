package domain

import (
	"fmt"
	"time"
)

// Enum that defines various states an auction can be in
type AuctionState string

const (
	PENDING   AuctionState = "PENDING"
	ACTIVE    AuctionState = "ACTIVE"
	CANCELED  AuctionState = "CANCELED"
	COMPLETED AuctionState = "COMPLETED"
	UNKNOWN   AuctionState = "UKNOWN"
)

type Auction struct {
	Item               *Item
	bids               []*Bid // slice of pointers to bids; new higher bids get appended on the end
	cancellation       *Cancellation
	sentStartSoonAlert bool
	sentEndSoonAlert   bool
	finalized          bool
}

func NewAuction(item *Item, bids []*Bid, cancellation *Cancellation, sentStartSoonAlert, sentEndSoonAlert, finalized bool) *Auction {
	return &Auction{
		Item:               item,
		bids:               bids,         // nil if brand new
		cancellation:       cancellation, // nil if brand new
		sentStartSoonAlert: sentEndSoonAlert,
		sentEndSoonAlert:   sentEndSoonAlert,
		finalized:          finalized,
	}
}

func (auction *Auction) ProcessNewBid(incomingBid *Bid) bool {
	timeBidReceived := incomingBid.TimeReceived
	stateWhenBidReceived := auction.getStateAtTime(timeBidReceived)

	switch {
	case stateWhenBidReceived == PENDING:
		fmt.Println("ignoring bid. auction hadn't begun when bid was received.")
	case stateWhenBidReceived == CANCELED:
		fmt.Println("ignoring bid. auction was cancelled before bid was received.")
	case stateWhenBidReceived == COMPLETED:
		fmt.Println("ignoring bid. auction was over before bid was received.")
	case stateWhenBidReceived == ACTIVE:
		highestBid := auction.GetHighestBid()
		if highestBid == nil { // case: first bid getting added
			if incomingBid.AmountInCents >= auction.Item.StartPriceInCents { // bid amount must at least be start price
				auction.addBid(incomingBid)
				auction.alertSeller("you have a new top bid!")
				return true
			}
		} else { // case: auction already has at least one bid
			if incomingBid.Outbids(highestBid) {
				auction.addBid(incomingBid)
				auction.alertSeller("you have a new top bid!")
				auction.alertBidder("your top bid has been out-matched!", highestBid)
				return true
			}
		}
	default:
		panic("see processNewBid()! couldn't process bid because didn't understand state of auction at time bid was received.")
	}
	return false
}

func (auction *Auction) addBid(bid *Bid) {
	auction.bids = append(auction.bids, bid)
}

func (auction *Auction) GetHighestBid() *Bid {
	if len(auction.bids) == 0 {
		return nil
	}
	return auction.bids[len(auction.bids)-1]
}

func (auction *Auction) getStateAtTime(currTime time.Time) AuctionState {
	// if auction has been cancelled, then if the time was
	// after the time of the cancellation, then the state at that
	// time is cancelled
	if auction.HasCancellation() {
		if AfterOrOn(&currTime, &auction.cancellation.TimeReceived) {
			return CANCELED
		}
	}

	// if time is before the auction start time, the auction is pending
	if currTime.Before(auction.Item.StartTime) {
		return PENDING
	}

	// if time is between the start and end time, inclusive, the auction is active
	atOrAfterStart := AfterOrOn(&currTime, &auction.Item.StartTime)
	atOrBeforeEnd := BeforeOrOn(&currTime, &auction.Item.EndTime)
	if atOrAfterStart && atOrBeforeEnd {
		return ACTIVE
	}

	// if time is after auction end time, the auction is completed
	// (already checked if auction has been cancelled)
	if currTime.After(auction.Item.EndTime) {
		return COMPLETED
	}

	panic("Auction.GetStateAtTime() couldn't determine auction state at time!")
}

func (auction *Auction) alertSeller(msg string) {
	sellerUserId := auction.Item.SellerUserId
	fmt.Printf("STUBBED: sending out request to notify seller (userId=%s,msg=%s)\n", sellerUserId, msg)
}

func (auction *Auction) alertBidder(msg string, bid *Bid) {
	bidderUserId := bid.BidderUserId
	fmt.Printf("STUBBED: sending out request to notify bidder (userId=%s,msg=%s)\n", bidderUserId, msg)
}

func (auction *Auction) Cancel(timeWhenCancellationIssued time.Time) bool {
	state := auction.getStateAtTime(timeWhenCancellationIssued)
	switch {
	case state == PENDING || state == ACTIVE:
		auction.cancellation = NewCancellation(timeWhenCancellationIssued)
		return true
	default:
		return false // state is COMPLETED, or CANCELED already
	}

}

func (auction *Auction) Stop(timeWhenStopIssued time.Time) bool {
	// this cancellation succeeds if before auction end time
	if BeforeOrOn(&timeWhenStopIssued, &auction.Item.EndTime) {
		auction.cancellation = NewCancellation(timeWhenStopIssued)
		return true
	}
	return false
}

func (auction *Auction) HasCancellation() bool {
	if auction.cancellation != nil {
		return true
	}
	return false
}

func (auction *Auction) IsOverOrCanceled() bool {
	if auction.cancellation != nil || time.Now().After(auction.Item.EndTime) {
		return true
	}
	return false
}

func (auction *Auction) hasBid(bidId string) bool {
	for _, bid := range auction.bids {
		if bid.BidId == bidId {
			return true
		}
	}
	return false
}

func (auction *Auction) OverlapsWith(leftBound *time.Time, rightBound *time.Time) bool {
	if rightBound.Before(auction.Item.StartTime) || leftBound.After(auction.Item.EndTime) {
		return false
	}
	return true
}

func AfterOrOn(someTime *time.Time, otherTime *time.Time) bool {
	return someTime.After(*otherTime) || someTime.Equal(*otherTime)
}

func BeforeOrOn(someTime *time.Time, otherTime *time.Time) bool {
	return someTime.Before(*otherTime) || someTime.Equal(*otherTime)
}

func (auction *Auction) SendStartSoonAlertIfApplicable() bool {
	if !auction.sentStartSoonAlert {
		fmt.Println("[Auction] sending out starting soon alert")
		return true
	}
	return false
}

func (auction *Auction) SendEndSoonAlertIfApplicable() bool {
	if !auction.sentStartSoonAlert {
		fmt.Println("[Auction] sending out ending soon alert")
		return true
	}
	return false
}

func (auction *Auction) Finalize() bool {
	fmt.Println("[Auction] finalizing myself")
	return true
}

func (auction *Auction) IsFinalized() bool {
	return auction.finalized
}
