package main

// this file is the entire interface of the auctions-service

import (
	"auctions-service/domain"
	"fmt"
	"sync"
	"time"
)

type AuctionService struct {
	bidRepo          domain.BidRepository
	auctionRepo      domain.AuctionRepository
	inMemoryAuctions map[string]*domain.Auction
	mutex            *sync.Mutex // for now, using coarse-grained concurrency implementation
}

func NewAuctionService(bidRepo domain.BidRepository, auctionRepo domain.AuctionRepository) *AuctionService {
	inMemoryAuctions := map[string]*domain.Auction{}
	mutex := &sync.Mutex{}
	return &AuctionService{
		bidRepo:          bidRepo,
		auctionRepo:      auctionRepo,
		inMemoryAuctions: inMemoryAuctions,
		mutex:            mutex,
	}
}

type AuctionInteractionOutcome string

const (
	auctionAlreadyCreated                   AuctionInteractionOutcome = "ALREADY_CREATED"      // create
	auctionSuccessfullyCreated              AuctionInteractionOutcome = "CREATED_SUCCESSFULLY" // create
	auctionWouldStartTooSoon                AuctionInteractionOutcome = "STARTS_TOO_SOON"      // create
	auctionStartsInPast                     AuctionInteractionOutcome = "STARTS_IN_PAST"       // create
	badTimeSpecified                        AuctionInteractionOutcome = "BAD_TIME_SPECIFIED_TIME"
	auctionSuccessfullyCanceled             AuctionInteractionOutcome = "CANCELED_SUCCESSFULLY"   // cancel
	auctionSuccessfullyStopped              AuctionInteractionOutcome = "STOPPED_SUCCESSFULLY"    // stop
	auctionNotExist                         AuctionInteractionOutcome = "AUCTION_NOT_EXIST"       // cancel, stop
	auctionAlreadyCanceled                  AuctionInteractionOutcome = "ALREADY_CANCELED"        // cancel
	auctionAlreadyOver                      AuctionInteractionOutcome = "ALREADY_OVER"            // cancel, stop
	auctionAlreadyFinalized                 AuctionInteractionOutcome = "ALREADY_FINALIZED"       // cancel, stop
	auctionCancellationRequesterIsNotSeller AuctionInteractionOutcome = "REQUESTER_IS_NOT_SELLER" // cancel
	auctionProcessedBid                     AuctionInteractionOutcome = "BID_WAS_SEEN_BY_AUCTION" // cancel
)

func (auctionservice *AuctionService) CreateAuction(itemId, sellerUserId string, startTime, endTime *time.Time, startPriceInCents int64) AuctionInteractionOutcome {

	auctionservice.mutex.Lock()

	fmt.Println("[AuctionService] creating Auction...")

	// confirm well-specified time
	if !endTime.After(*startTime) {
		return badTimeSpecified
	}

	creationTime := time.Now()

	// confirm auction does not start in the past
	if creationTime.After(*startTime) {
		return auctionStartsInPast
	}

	// if auction to be created will start in sooner than 2 hours, do not proceed
	if creationTime.Add(time.Duration(2) * time.Hour).After(*startTime) {
		return auctionWouldStartTooSoon
	}

	// confirm an auction hasn't already been created for the item
	if auctionservice.auctionRepo.GetAuction(itemId) != nil {
		return auctionAlreadyCreated
	}

	newItem := domain.NewItem(itemId, sellerUserId, *startTime, *endTime, startPriceInCents)
	newAuction := domain.NewAuction(newItem, nil, nil, false, false, nil)

	auctionservice.auctionRepo.SaveAuction(newAuction)                   // save Auction
	auctionservice.inMemoryAuctions[newAuction.Item.ItemId] = newAuction // cache Auction
	auctionservice.mutex.Unlock()
	// auctionservice.addAuction(newAuction)
	// auctionservice.auctionRepo.SaveAuction()
	return auctionSuccessfullyCreated
}

func (auctionservice *AuctionService) CancelAuction(itemId string, requesterUserId string) AuctionInteractionOutcome {

	auctionservice.mutex.Lock()

	fmt.Println("[AuctionService] cancelling Auction...")
	timeWhenCancelReceived := time.Now()

	relevantAuction, ok := auctionservice.inMemoryAuctions[itemId] // lookup in cache
	if !ok {
		relevantAuction = auctionservice.auctionRepo.GetAuction(itemId) // get from db if not cached
	} // dont bother caching though

	// confirm auction exists
	if relevantAuction == nil {
		return auctionNotExist
	}

	// confirm the person requesting an auction be canceled is the seller of the item
	if relevantAuction.Item.SellerUserId != requesterUserId {
		return auctionCancellationRequesterIsNotSeller
	}

	// confirm auction isn't already finalized
	if relevantAuction.HasFinalization() {
		return auctionAlreadyFinalized
	}

	// confirm auction isn't already canceled
	if relevantAuction.HasCancellation() {
		return auctionAlreadyCanceled
	}

	// confirm auction isn't already over (at time)
	if relevantAuction.IsOverOrCanceledAtTime(timeWhenCancelReceived) {
		return auctionAlreadyOver
	}

	// otherwise, should be ok to cancel.

	wasCanceled := relevantAuction.Cancel(timeWhenCancelReceived) // should always return true...
	if wasCanceled {
		auctionservice.auctionRepo.SaveAuction(relevantAuction) // save Auction
		// dont cache Auction
	}

	auctionservice.mutex.Unlock()

	if wasCanceled {
		return auctionSuccessfullyCanceled
	} else {
		panic("[AuctionService] see CancelAuction(). reached end of method without determining what happened (bug).")
	}
}

func (auctionservice *AuctionService) StopAuction(itemId string) AuctionInteractionOutcome {

	auctionservice.mutex.Lock()

	fmt.Println("[AuctionService] stopping Auction.")
	timeWhenStopReceived := time.Now()

	relevantAuction, ok := auctionservice.inMemoryAuctions[itemId] // lookup in cache
	toCache := false
	if !ok {
		relevantAuction = auctionservice.auctionRepo.GetAuction(itemId) // get from db if not cached
		toCache = true
	} // cache it if successful b/c it means we will need to finalized it.

	// confirm auction exists
	if relevantAuction == nil {
		return auctionNotExist
	}

	// assume client code confirmed requester is an admin

	// confirm auction isn't already finalized
	if relevantAuction.HasFinalization() {
		return auctionAlreadyFinalized
	}

	// confirm auction isn't already canceled
	if relevantAuction.HasCancellation() {
		return auctionAlreadyCanceled
	}

	// confirm auction isn't already over
	if relevantAuction.IsOverOrCanceledAtTime(timeWhenStopReceived) {
		return auctionAlreadyOver
	}

	// otherwise, ok to stop.
	wasStopped := relevantAuction.Cancel(time.Now()) // should always return true?
	if wasStopped {
		auctionservice.auctionRepo.SaveAuction(relevantAuction)
	}

	auctionservice.mutex.Unlock()

	if wasStopped {
		if toCache {
			auctionservice.inMemoryAuctions[itemId] = relevantAuction // cache it
		}
		return auctionSuccessfullyStopped
	} else {
		panic("[AuctionService] see StopAuction(). reached end of method without determining what happened (bug).")
	}

}

// type BidInteractionOutcome string

// const (
// 	successfullyNewTopBid                   BidInteractionOutcome = "SUCCESSFULLY_NEW_TOP_BID"      // process new bid
// 	associatedAuctionDoesNotExist
// )

func (auctionservice *AuctionService) ProcessNewBid(itemId string, bidderUserId string, timeReceived time.Time, amountInCents int64) (AuctionInteractionOutcome, domain.AuctionState, bool) {

	auctionservice.mutex.Lock()

	fmt.Println("[AuctionService] processing new bid.")
	newId := auctionservice.bidRepo.NextBidId()
	newBid := domain.NewBid(newId, itemId, bidderUserId, timeReceived, amountInCents, true)

	relevantAuction, ok := auctionservice.inMemoryAuctions[itemId] // lookup in cache
	toCache := false
	if !ok {
		relevantAuction = auctionservice.auctionRepo.GetAuction(itemId) // get from db if not cached
		if relevantAuction == nil {
			return auctionNotExist, domain.UNKNOWN, false // unknown auction state == auction not exist
		}
		toCache = true
	} // cache the auction if the bid ends up successfully being placed.

	auctionState, wasNewTopBid := relevantAuction.ProcessNewBid(newBid)
	if wasNewTopBid {
		auctionservice.bidRepo.SaveBid(newBid) // only save bids that were determined to be new Top bids
		if toCache {
			auctionservice.inMemoryAuctions[itemId] = relevantAuction
		}
	}

	auctionservice.mutex.Unlock()

	return auctionProcessedBid, auctionState, wasNewTopBid

}

func (auctionservice *AuctionService) GetItemsUserHasBidsOn(userId string) *[]string {
	fmt.Println("[AuctionService] getting and returning items.")
	bids := auctionservice.bidRepo.GetBidsByUserId(userId) // includes inactive bids
	itemIds := make([]string, 0)
	alreadySeenItemIds := map[string]interface{}{}
	for _, bid := range *bids {
		if _, ok := alreadySeenItemIds[bid.ItemId]; !ok {
			itemIds = append(itemIds, bid.ItemId)
			alreadySeenItemIds[bid.ItemId] = nil
		}
	}
	return &itemIds
}

func (auctionservice *AuctionService) GetActiveAuctions() *[]*domain.Auction {
	fmt.Println("[AuctionService] getting and returning active auctions.")
	now := time.Now()
	auctions := auctionservice.auctionRepo.GetAuctions(now, now)
	return &auctions
}

func (auctionservice *AuctionService) ActivateUserBids(userId string) (int, int) {

	auctionservice.mutex.Lock()

	fmt.Printf("[AuctionService] activating user %s bids.", userId)

	timeWhenUserActivated := time.Now()

	userBids := auctionservice.bidRepo.GetBidsByUserId(userId)
	itemIds := make([]string, 0) // list of all items (auctions) the user has bids in
	alreadySeenItemIds := map[string]interface{}{}
	for _, bid := range *userBids {
		if _, ok := alreadySeenItemIds[bid.ItemId]; !ok {
			itemIds = append(itemIds, bid.ItemId)
			alreadySeenItemIds[bid.ItemId] = nil
		}
	}

	bidsToUpdateInRepo := []*domain.Bid{}
	numAuctionsWBidUpdates := 0

	for _, itemId := range itemIds {
		auction, ok := auctionservice.inMemoryAuctions[itemId]
		if !ok { // did not find auction in memory; bring into memory
			auction = auctionservice.auctionRepo.GetAuction(itemId)
		}
		bidsToSave, _ := auction.ActivateUserBids(userId, timeWhenUserActivated) // returns the bids whose state was changed
		bidsToUpdateInRepo = append(bidsToUpdateInRepo, *bidsToSave...)
		numAuctionsWBidUpdates++
	}

	for _, bid := range bidsToUpdateInRepo {
		auctionservice.bidRepo.SaveBid(bid)
	}

	auctionservice.mutex.Unlock()

	return len(bidsToUpdateInRepo), numAuctionsWBidUpdates

}

func (auctionservice *AuctionService) DeactivateUserBids(userId string) (int, int) {

	auctionservice.mutex.Lock()

	fmt.Printf("[AuctionService] de-activating user %s bids.", userId)

	timeWhenUserDeactivated := time.Now()

	userBids := auctionservice.bidRepo.GetBidsByUserId(userId)
	itemIds := make([]string, 0) // list of all items (auctions) the user has bids in
	alreadySeenItemIds := map[string]interface{}{}
	for _, bid := range *userBids {
		if _, ok := alreadySeenItemIds[bid.ItemId]; !ok {
			itemIds = append(itemIds, bid.ItemId)
			alreadySeenItemIds[bid.ItemId] = nil
		}
	}

	bidsToUpdateInRepo := []*domain.Bid{}
	numAuctionsWBidUpdates := 0

	for _, itemId := range itemIds {
		auction, ok := auctionservice.inMemoryAuctions[itemId]
		if !ok { // did not find auction in memory; bring into memory
			auction = auctionservice.auctionRepo.GetAuction(itemId)
		}
		bidsToSave, _ := auction.DeactivateUserBids(userId, timeWhenUserDeactivated) // returns the bids whose state was changed
		bidsToUpdateInRepo = append(bidsToUpdateInRepo, *bidsToSave...)
		numAuctionsWBidUpdates++
	}

	for _, bid := range bidsToUpdateInRepo {
		auctionservice.bidRepo.SaveBid(bid)
	}

	auctionservice.mutex.Unlock()

	return len(bidsToUpdateInRepo), numAuctionsWBidUpdates
}

func (auctionservice *AuctionService) LoadAuctionsIntoMemory(sinceTime time.Time, upToTime time.Time) {

	inMemAuctions := auctionservice.inMemoryAuctions

	auctionservice.mutex.Lock()

	fmt.Println("[AuctionService] loading auctions into memory")

	auctions := auctionservice.auctionRepo.GetAuctions(sinceTime, upToTime)
	for _, auction := range auctions {
		if !auction.HasFinalization() { // dont bring into memory if it is a finalized auction
			if _, ok := (inMemAuctions)[auction.Item.ItemId]; !ok {
				(inMemAuctions)[auction.Item.ItemId] = auction
			}
		}

	}
	auctionservice.mutex.Unlock()
}

func (auctionservice *AuctionService) SendOutLifeCycleAlerts() {
	inMemAuctions := auctionservice.inMemoryAuctions
	var sentNotif1, sentNotif2 bool

	auctionservice.mutex.Lock()

	fmt.Println("[AuctionService] sending out life cycle alerts")

	for _, auction := range inMemAuctions {
		sentNotif1 = auction.SendStartSoonAlertIfApplicable()
		sentNotif2 = auction.SendEndSoonAlertIfApplicable()
		if sentNotif1 || sentNotif2 {
			auctionservice.auctionRepo.SaveAuction(auction) // save the knowledge that alert was sent out;
		}
	}

	auctionservice.mutex.Unlock()
}

func (auctionservice *AuctionService) FinalizeAnyPastAuctions(finalizeDelay time.Duration) {
	inMemAuctions := auctionservice.inMemoryAuctions

	auctionservice.mutex.Lock()

	for _, auction := range inMemAuctions {
		wasFinalized := auction.Finalize(time.Now())
		if wasFinalized {
			auctionservice.auctionRepo.SaveAuction(auction) // save the knowledge that we finalized the auction
		}
	}

	auctionservice.mutex.Unlock()
}
