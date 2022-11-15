package domain

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestGetHighestBidder(t *testing.T) {
	startime := time.Date(2014, 2, 4, 01, 00, 00, 0, time.UTC)
	endtime := time.Date(2014, 2, 4, 01, 30, 00, 0, time.UTC)            // 30 min later
	item := NewItem("101", "asclark109", startime, endtime, int64(2000)) // $20 start price
	auction1 := NewAuction(item, nil, nil, false, false, false)

	timeBidsReceived := startime.Add(time.Duration(15) * time.Second) // 15 seconds into auctions tart

	amounts := []int64{400, 12, 32545, 453, 351, 534, 684, 651, 6, 516}
	var bids []*Bid = make([]*Bid, len(amounts))

	var maxBid Bid
	for i := 0; i < len(amounts); i++ {
		bids[i] = NewBid("40", "101", "104", timeBidsReceived, amounts[i], true)
		if i == 0 {
			maxBid = *bids[0]
		} else {
			if bids[i].AmountInCents > maxBid.AmountInCents {
				maxBid = *bids[i]
			}
		}
	}

	for _, bid := range bids {
		auction1.addBid(bid)
	}

	result := auction1.GetHighestBid().BidId
	expected := maxBid.BidId

	if result != expected {
		t.Errorf("\nRan:%s\nExpected:%s\nGot:%s", "auction.GetHighestBid()", expected, result)
	}

}

func TestAddBidHasBid(t *testing.T) {
	startime := time.Date(2014, 2, 4, 01, 00, 00, 0, time.UTC)
	endtime := time.Date(2014, 2, 4, 01, 30, 00, 0, time.UTC)            // 30 min later
	item := NewItem("101", "asclark109", startime, endtime, int64(2000)) // $20 start price
	auction1 := NewAuction(item, nil, nil, false, false, false)

	timeBid1Received := startime.Add(time.Duration(15) * time.Minute) // 10 min after auction start
	timeBid2Received := startime.Add(time.Duration(16) * time.Minute) // 11 min after auction start
	bid1 := NewBid("40", "101", "104", timeBid1Received, 4000, true)  // $40
	bid2 := NewBid("40", "101", "104", timeBid2Received, 5000, true)

	// confirm auction doesn't have the bid
	result := auction1.hasBid(bid1.BidId)
	expected := false
	if result != expected {
		t.Errorf("\nRan:%s\nExpected:%s\nGot:%s", "auction.hasBid(bidId string)", strconv.FormatBool(expected), strconv.FormatBool(result))
	}

	// add bid1
	auction1.addBid(bid1)

	// confirm auction has bid1
	result = auction1.hasBid(bid1.BidId)
	expected = true
	if result != expected {
		t.Errorf("\nRan:%s\nExpected:%s\nGot:%s", "auction.hasBid(bidId string)", strconv.FormatBool(expected), strconv.FormatBool(result))
	}

	// add bid2
	auction1.addBid(bid1)

	// confirm auction has bid1 and bid2
	result = auction1.hasBid(bid1.BidId) && auction1.hasBid(bid2.BidId)
	expected = true
	if result != expected {
		t.Errorf("\nRan:%s\nExpected:%s\nGot:%s", "auction.hasBid(bidId string)", strconv.FormatBool(expected), strconv.FormatBool(result))
	}

}

func TestCancelAuction(t *testing.T) {
	startime := time.Date(2014, 2, 4, 01, 00, 00, 0, time.UTC)
	endtime := time.Date(2014, 2, 4, 01, 30, 00, 0, time.UTC)            // 30 min later
	item := NewItem("101", "asclark109", startime, endtime, int64(2000)) // $20 start price
	auction1 := NewAuction(item, nil, nil, false, false, false)
	auction2 := NewAuction(item, nil, nil, false, false, false)
	auction3 := NewAuction(item, nil, nil, false, false, false)
	auction4 := NewAuction(item, nil, nil, false, false, false) // have this auction already be canceled
	auction4.Cancel(startime)

	canceltime1 := time.Date(2014, 2, 4, 00, 00, 00, 0, time.UTC) // before auction starts
	canceltime2 := time.Date(2014, 2, 4, 01, 10, 00, 0, time.UTC) // while auction going
	canceltime3 := time.Date(2014, 2, 4, 01, 40, 00, 0, time.UTC) // after auction over

	var tests = []struct {
		auction    *Auction
		cancelTime time.Time
		expected   bool
	}{
		{auction1, canceltime1, true},  // can cancel if auction hasn't started (pending)
		{auction2, canceltime2, true},  // can cancel if auction is active (live)
		{auction3, canceltime3, false}, // can't cancel after auction over (completed)
		{auction4, canceltime2, false}, // can't cancel auction if already canceled
	}

	for num, test := range tests {
		auction := test.auction
		cancelTime := test.cancelTime
		testname := fmt.Sprintf("T=%v", num)
		t.Run(testname, func(t *testing.T) {
			result := auction.Cancel(cancelTime)
			if result != test.expected {
				t.Errorf("\nRan:%s\nExpected:%s\nGot:%s", "auction.Cancel()", strconv.FormatBool(test.expected), strconv.FormatBool(result))
			}
		})
	}

}

func TestGetStateAtTime(t *testing.T) {

	startime := time.Date(2014, 2, 4, 01, 00, 00, 0, time.UTC)
	endtime := time.Date(2014, 2, 4, 01, 30, 00, 0, time.UTC)            // 30 min later
	item := NewItem("101", "asclark109", startime, endtime, int64(2000)) // $20 start price

	auction := NewAuction(item, nil, nil, false, false, false) // uncanceled auction

	auction_cancelled := NewAuction(item, nil, nil, false, false, false) // auction cancelled 15 min into auction start (30 min long auction)
	canceltime := time.Date(2014, 2, 4, 01, 15, 00, 0, time.UTC)
	auction_cancelled.Cancel(canceltime)

	time1 := time.Date(2014, 2, 4, 00, 30, 00, 0, time.UTC)     // 30 min before auction start
	time2 := startime.Add(-time.Duration(1) * time.Microsecond) // 1 microsecond before start
	time3 := startime                                           // on auction start
	time4 := startime.Add(time.Duration(15) * time.Minute)      // halfway thru auction
	time5 := endtime                                            // on auction end
	time6 := endtime.Add(time.Duration(time.Microsecond))       // 1 microsecond after end

	// fmt.Println("time1: ", time1)
	// fmt.Println("time2: ", time2)
	// fmt.Println("time3: ", time3)
	// fmt.Println("time4: ", time4)
	// fmt.Println("time5: ", time5)
	// fmt.Println("time6: ", time6)

	var tests = []struct {
		auction  *Auction
		currTime time.Time
		expected AuctionState
	}{
		{auction, time1, PENDING},
		{auction, time2, PENDING},
		{auction, time3, ACTIVE},
		{auction, time4, ACTIVE},
		{auction, time5, ACTIVE},
		{auction, time6, COMPLETED},
		{auction_cancelled, time1, PENDING},
		{auction_cancelled, time2, PENDING},
		{auction_cancelled, time3, ACTIVE},
		{auction_cancelled, time4, CANCELED},
		{auction_cancelled, time5, CANCELED},
		{auction_cancelled, time6, CANCELED},
	}

	for num, test := range tests {
		auction := test.auction
		currTime := test.currTime
		expectedState := test.expected
		testname := fmt.Sprintf("T=%v", num)
		t.Run(testname, func(t *testing.T) {
			calcdState := auction.getStateAtTime(currTime)
			if calcdState != expectedState {
				t.Errorf("\nRan:%s\nExpected:%s\nGot:%s", "auction.GetStateAtTime(currTime)", expectedState, calcdState)
			}
		})
	}
}

func TestProcessBid(t *testing.T) {
	startime := time.Date(2014, 2, 4, 01, 00, 00, 0, time.UTC)
	endtime := time.Date(2014, 2, 4, 01, 30, 00, 0, time.UTC)             // 30 min later
	item1 := NewItem("101", "asclark109", startime, endtime, int64(2000)) // $20 start price
	item2 := NewItem("102", "asclark109", startime, endtime, int64(2000)) // $20 start price
	auction1 := NewAuction(item1, nil, nil, false, false, false)          // will go to completion
	auction2 := NewAuction(item2, nil, nil, false, false, false)          // will get cancelled halfway through

	time1 := startime.Add(-time.Duration(30) * time.Minute)     // 30 min before auction start;   $0.12; ignored
	time2 := startime.Add(-time.Duration(1) * time.Microsecond) // 1 microsecond before start;  $400.00; ignored
	time3 := startime                                           // on auction start;              $0.12; ignored because start price too low
	time4 := startime.Add(time.Duration(1) * time.Microsecond)  // 1 microsecond later;          $20.00; new top bid
	time5 := time4.Add(time.Duration(1) * time.Minute)          // 1 minute later;               $25.00; new top bid
	time6 := time5.Add(time.Duration(1) * time.Minute)          // 1 minute later;               $22.25; ignored because amount too low
	time7 := endtime                                            // exactly at the end;           $30.00; new top bid
	time8 := endtime.Add(time.Duration(1) * time.Microsecond)   // 1 microsecond after end;     $300.00; ignored because after auction end

	amount1 := int64(12)
	amount2 := int64(40000)
	amount3 := int64(12)
	amount4 := int64(2000)
	amount5 := int64(2500)
	amount6 := int64(2225)
	amount7 := int64(3000)
	amount8 := int64(30000)

	times := []time.Time{
		time1,
		time2,
		time3,
		time4,
		time5,
		time6,
		time7,
		time8,
	}

	amounts := []int64{
		amount1,
		amount2,
		amount3,
		amount4,
		amount5,
		amount6,
		amount7,
		amount8,
	}

	var bidsForAuction1 []*Bid = make([]*Bid, len(amounts))
	var bidsForAuction2 []*Bid = make([]*Bid, len(amounts))

	auction2.Cancel(time6) // cancel auction 2 after a few bids have come in;
	// this should cause the final bid comming in when the auction ends to be ignored;

	fmt.Println("Auction 1 processing bids")
	idx := 1
	for i := 0; i < len(amounts); i++ {
		bidsForAuction1[i] = NewBid(fmt.Sprint(idx), "101", "asclark", times[i], amounts[i], true)
		auction1.ProcessNewBid(bidsForAuction1[i])
		idx++
	}

	fmt.Println("Auction 2 processing bids")
	idx = 1
	for i := 0; i < len(amounts); i++ {
		bidsForAuction2[i] = NewBid(fmt.Sprint(idx), "102", "asclark", times[i], amounts[i], true)
		auction2.ProcessNewBid(bidsForAuction2[i])
		idx++
	}

	// expect the very last bid placed in auction1 (bid7) to be highest (legal) Bid
	result := auction1.GetHighestBid().BidId
	expected := "7"
	if result != expected {
		t.Errorf("\nRan:%s\nExpected:%s\nGot:%s", "auction.GetStateAtTime(currTime)", expected, result)
	}

	// in Auction2 we don't really care about who the highest bid is because it was cancelled
	// but to confirm this Auction ignored incoming bids after it was cancelled, confirm that
	// bid5 was the last bid placed (i.e. the highest bid)
	// expect the very last bid placed in auction1 (bid7) to be highest (legal) Bid
	result = auction2.GetHighestBid().BidId
	expected = "5"
	if result != expected {
		t.Errorf("\nRan:%s\nExpected:%s\nGot:%s", "auction.GetStateAtTime(currTime)", expected, result)
	}

}
