package domain

import (
	"fmt"
	"time"
)

// Enum that defines various states an auction can be in
type AuctionState string

const (
	PENDING   AuctionState = "PENDING" // has not yet started
	ACTIVE    AuctionState = "ACTIVE"  // is happening now
	CANCELED  AuctionState = "CANCELED"
	OVER      AuctionState = "OVER"      // is over (but winner has not been declared and auction has not been "archived away")
	FINALIZED AuctionState = "FINALIZED" // is over and archived away; can delete
	UNKNOWN   AuctionState = "UKNOWN"
)

type Auction struct {
	Item               *Item
	bids               []*Bid // slice of pointers to bids; new higher bids get appended on the end
	cancellation       *Cancellation
	sentStartSoonAlert bool
	sentEndSoonAlert   bool
	finalization       *Finalization
}

func NewAuction(item *Item, bids []*Bid, cancellation *Cancellation, sentStartSoonAlert, sentEndSoonAlert bool, finalization *Finalization) *Auction {
	return &Auction{
		Item:               item,
		bids:               bids,             // nil if brand new
		cancellation:       cancellation,     // nil if brand new
		sentStartSoonAlert: sentEndSoonAlert, // false if brand new
		sentEndSoonAlert:   sentEndSoonAlert, // false if brand new
		finalization:       finalization,     // nil if brand new
	}
}

func (auction *Auction) ProcessNewBid(incomingBid *Bid) (AuctionState, bool) {
	timeBidReceived := incomingBid.TimeReceived
	stateWhenBidReceived := auction.getStateAtTime(timeBidReceived)

	// if the auction has been finalized, it is archived and we are no longer
	// considering new bids.
	if auction.HasFinalization() { // i.e. has state FINALIZED at some known point in time
		fmt.Println("[Auction] ignoring bid. auction has been finalized.")
		return FINALIZED, false
	}

	switch {
	case stateWhenBidReceived == PENDING:
		fmt.Println("[Auction] ignoring bid. auction hadn't begun when bid was received.")
		return PENDING, false
	case stateWhenBidReceived == CANCELED:
		fmt.Println("[Auction] ignoring bid. auction was cancelled before bid was received.")
		return CANCELED, false
	case stateWhenBidReceived == OVER:
		fmt.Println("[Auction] ignoring bid. auction was over before bid was received.")
		return OVER, false
	// case stateWhenBidReceived == FINALIZED: HANDLED ABOVE
	case stateWhenBidReceived == ACTIVE:
		highestActiveBid := auction.GetHighestActiveBid()
		if highestActiveBid == nil { // case: there are no active bids
			if incomingBid.AmountInCents >= auction.Item.StartPriceInCents { // bid amount must at least be start price
				fmt.Println("[Auction] new top bid!")
				auction.addBid(incomingBid)
				auction.alertSeller("you have a new top bid!")
				return ACTIVE, true
			} else {
				fmt.Println("[Auction] ignoring bid. bid was under start price.")
				return ACTIVE, false
			}
		} else { // case: auction already has at least one active bid
			if incomingBid.Outbids(highestActiveBid) {
				fmt.Println("[Auction] new top bid!")
				auction.addBid(incomingBid)
				auction.alertSeller("you have a new top bid!")
				auction.alertBidder("your top bid has been out-matched!", highestActiveBid)
				return ACTIVE, true
			} else {
				fmt.Println("[Auction] ignoring bid. bid was under highest bid offer amount.")
				return ACTIVE, false
			}
		}
	default:
		panic("see processNewBid()! couldn't process bid because didn't understand state of auction at time bid was received.")
	}
}

func (auction *Auction) addBid(bid *Bid) {
	auction.bids = append(auction.bids, bid)
}

func (auction *Auction) GetHighestActiveBid() *Bid {
	if len(auction.bids) == 0 {
		return nil
	}
	// by convention, top bids are appended at the end; so start at end and walk to the left.
	// find the first active bid
	idx := len(auction.bids) - 1
	for !auction.bids[idx].active {
		idx--
		if idx == -1 {
			return nil
		}
	}
	return auction.bids[idx]
}

func (auction *Auction) getStateAtTime(currTime time.Time) AuctionState {
	// if auction has been cancelled, then if the time was
	// after the time of the cancellation, then the state at that
	// time is cancelled
	if auction.HasFinalization() {
		if AfterOrOn(&currTime, &auction.finalization.TimeReceived) {
			return FINALIZED
		}
	}

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

	// if time is after auction end time (auction has not been cancelled nor finalized), then the auction is over
	// (already checked if auction has been cancelled)
	if currTime.After(auction.Item.EndTime) {
		return OVER
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

	// cant issue cancel if there is already a cancellation, or the auction is considered finalized
	if auction.HasCancellation() || auction.HasFinalization() {
		return false // only allow 1 cancellation; don't allow any changes once finalized
	}

	// otherwise, the auction is pending, active, or over.
	// can only issue cancel if auction is pending or active and has no bids
	stateWhenCancellationIssued := auction.getStateAtTime(timeWhenCancellationIssued)
	switch {
	case stateWhenCancellationIssued == PENDING: //
		auction.cancellation = NewCancellation(timeWhenCancellationIssued)
		return true
	case stateWhenCancellationIssued == ACTIVE && !auction.HasActiveBid(): //
		auction.cancellation = NewCancellation(timeWhenCancellationIssued)
		return true
	default:
		return false // state is COMPLETED, or CANCELED already
	}
}

func (auction *Auction) HasActiveBid() bool {
	for _, bid := range auction.bids {
		if bid.active {
			return true
		}
	}
	return false
}

func (auction *Auction) Stop(timeWhenStopIssued time.Time) bool {

	// cant issue stop if there is already a cancellation, or the auction is considered finalized
	if auction.HasCancellation() || auction.HasFinalization() {
		return false // only allow 1 cancellation; don't allow any changes once finalized
	}

	// otherwise, the auction is pending, active, or over.
	// can only issue stop if auction is pending or active
	stateWhenStopIssued := auction.getStateAtTime(timeWhenStopIssued)
	switch {
	case stateWhenStopIssued == PENDING || stateWhenStopIssued == ACTIVE:
		auction.cancellation = NewCancellation(timeWhenStopIssued)
		return true
	default:
		return false // state is OVER, CANCELED, FINALIZED; cant stop
	}
}

func (auction *Auction) HasCancellation() bool {
	return auction.cancellation != nil
}

func (auction *Auction) HasFinalization() bool {
	return auction.finalization != nil
}

func (auction *Auction) DeactivateUserBids(userId string, timeWhenUserDeactivated time.Time) (*[]*Bid, bool) {
	// note: this call will deactivate all of the user's bids in the auction even
	// bids that are placed after the timeWhenUserDeactivated. timeWhenUserDeactivated
	// is only used to determine whether the auction outcome was "set-in-stone" when
	// the request to deactivate user's bids comes in; this is the only situation where
	// we refused a deactivateUserBids request
	stateWhenUserDeactivated := auction.getStateAtTime(timeWhenUserDeactivated)
	bidsToSave := []*Bid{}
	if stateWhenUserDeactivated == FINALIZED {
		return &bidsToSave, false
	} else {
		for _, bid := range auction.bids {
			if bid.BidderUserId == userId {
				gotDeactivated := bid.Deactivate()
				if gotDeactivated {
					bidsToSave = append(bidsToSave, bid)
				}
			}
		}
		return &bidsToSave, true
	}
}

func (auction *Auction) ActivateUserBids(userId string, timeWhenUserActivated time.Time) (*[]*Bid, bool) {
	stateWhenUserActivated := auction.getStateAtTime(timeWhenUserActivated)
	bidsToSave := []*Bid{}
	if stateWhenUserActivated == FINALIZED {
		return &bidsToSave, false
	} else {
		for _, bid := range auction.bids {
			if bid.BidderUserId == userId {
				wasActivated := bid.Activate()
				if wasActivated {
					bidsToSave = append(bidsToSave, bid)
				}
			}
		}
		return &bidsToSave, true
	}
}

func (auction *Auction) IsOverOrCanceledAtTime(atTime time.Time) bool {
	stateAtTime := auction.getStateAtTime(atTime)
	if stateAtTime == OVER || stateAtTime == CANCELED {
		return true
	}
	return false
}

func (auction *Auction) Finalize(timeWhenFinalizationIssued time.Time) bool {
	fmt.Println("[Auction] STUBBED finalizing self...")

	// cant issue finalization if this auction has already been finalized
	if auction.HasFinalization() {
		return false // only allow 1 finalization
	}

	// finalization only allowed when auction is canceled or over
	state := auction.getStateAtTime(timeWhenFinalizationIssued)
	switch {
	case state == CANCELED || state == OVER:
		auction.finalization = NewFinalization(timeWhenFinalizationIssued)
		return true
	default:
		return false // state is PENDING, ACTIVE, FINALIZED
	}

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

func (auction *Auction) SendStartSoonAlertIfApplicable() bool {
	if !auction.sentStartSoonAlert {
		fmt.Println("[Auction] sending out starting soon alert")
		return true
	}
	return false
}

func (auction *Auction) SendEndSoonAlertIfApplicable() bool {
	if !auction.sentEndSoonAlert {
		fmt.Println("[Auction] sending out ending soon alert")
		return true
	}
	return false
}

// misc helper functions

func AfterOrOn(someTime *time.Time, otherTime *time.Time) bool {
	return someTime.After(*otherTime) || someTime.Equal(*otherTime)
}

func BeforeOrOn(someTime *time.Time, otherTime *time.Time) bool {
	return someTime.Before(*otherTime) || someTime.Equal(*otherTime)
}
