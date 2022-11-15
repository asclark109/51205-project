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
	mutex            *sync.Mutex // for now, using coarse-grained concurrent hashtable
	// auctionSessionManager
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
	auctionCancellationRequesterIsNotSeller AuctionInteractionOutcome = "REQUESTER_IS_NOT_SELLER" // cancel
)

func (auctionservice *AuctionService) CreateAuction(itemId, sellerUserId string, startTime, endTime *time.Time, startPriceInCents int64) AuctionInteractionOutcome {
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
	newAuction := domain.NewAuction(newItem, nil, nil, false, false, false)

	auctionservice.mutex.Lock()
	auctionservice.auctionRepo.SaveAuction(newAuction)                   // save Auction
	auctionservice.inMemoryAuctions[newAuction.Item.ItemId] = newAuction // cache Auction
	auctionservice.mutex.Unlock()
	// auctionservice.addAuction(newAuction)
	// auctionservice.auctionRepo.SaveAuction()
	return auctionSuccessfullyCreated
}

func (auctionservice *AuctionService) CancelAuction(itemId string, requesterUserId string) AuctionInteractionOutcome {
	fmt.Println("[AuctionService] cancelling Auction...")

	auctionservice.mutex.Lock()

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

	// confirm auction isn't already canceled
	if relevantAuction.HasCancellation() {
		return auctionAlreadyCanceled
	}

	// confirm auction isn't already over
	if relevantAuction.IsOverOrCanceled() {
		return auctionAlreadyCanceled
	}

	// otherwise, ok to cancel.

	_ = relevantAuction.Cancel(time.Now())                  // should always return true? Confirm Later!!
	auctionservice.auctionRepo.SaveAuction(relevantAuction) // save Auction
	// dont cache Auction
	auctionservice.mutex.Unlock()

	return auctionSuccessfullyCanceled

}

func (auctionservice *AuctionService) StopAuction(itemId string) AuctionInteractionOutcome {
	fmt.Println("[AuctionService] stopping Auction.")

	auctionservice.mutex.Lock()

	relevantAuction, ok := auctionservice.inMemoryAuctions[itemId] // lookup in cache
	if !ok {
		relevantAuction = auctionservice.auctionRepo.GetAuction(itemId) // get from db if not cached
	} // dont bother caching though

	// confirm auction exists
	if relevantAuction == nil {
		return auctionNotExist
	}

	// assume client code confirmed requester is an admin

	// confirm auction isn't already canceled
	if relevantAuction.HasCancellation() {
		return auctionAlreadyCanceled
	}

	// confirm auction isn't already over
	if relevantAuction.IsOverOrCanceled() {
		return auctionAlreadyOver
	}

	// otherwise, ok to stop.
	_ = relevantAuction.Cancel(time.Now()) // should always return true?
	auctionservice.auctionRepo.SaveAuction(relevantAuction)
	auctionservice.mutex.Unlock()

	return auctionSuccessfullyStopped
}

func (auctionservice *AuctionService) ProcessNewBid() {
	fmt.Println("[AuctionService] processing new bid.")
}

func (auctionservice *AuctionService) GetItemsUserHasBidsOn(userId string) *[]string {
	fmt.Println("[AuctionService] getting and returning items.")
	bids := auctionservice.bidRepo.GetBidsByUserId(userId)
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

func (auctionservice *AuctionService) ActivateUserBids() {
	fmt.Println("[AuctionService] activating user's bids.")
}

func (auctionservice *AuctionService) DeactivateUserBids() {
	fmt.Println("[AuctionService] deactivating user's bids.")
}
