package main

import (
	"auctions-service/domain"
	"sync"
	"time"
)

const (
	loadAheadDuration  time.Duration = time.Duration(2) * time.Hour // how much in advance AuctionSessionManager should bring Auctions into memory before their start
	loadBehindDuration time.Duration = time.Duration(2) * time.Hour // how much in the past AuctionSessionManager should load Auctions into memory since their end
	// note, with a system shutdown and restart, AuctionSessionManager may have to finalize Auctions that ended e.g. 30 minutes ago
	FinalizeDelay time.Duration = time.Duration(30) * time.Minute // how long after auction end before SessionManager should finalize the auction
	// note, with a system shutdown and restart, the context may want to process bids sitting in queue that were submitted during live auction
	// session. so, give this context 30 minutes to get up to speed with its bid processing, which may involve processing bids as if from an
	// earlier point in time.
)

type AuctionSessionManager struct {
	inMemoryAuctions  *map[string]*domain.Auction
	mutex             *sync.Mutex
	auctionRepository domain.AuctionRepository
	turnedOn          bool
	lastAlertTime     time.Time
	lastFinalizeTime  time.Time
	lastLoadTime      time.Time
	alertCycle        time.Duration
	finalizeCycle     time.Duration
	loadCycle         time.Duration
}

func NewAuctionSessionManager(
	inMemoryAuctions *map[string]*domain.Auction,
	mutex *sync.Mutex,
	auctionRepo domain.AuctionRepository,
	alertCycle time.Duration,
	finalizeCycle time.Duration,
	loadAuctionCycle time.Duration,
) *AuctionSessionManager {

	lastAlertTime := time.Now()
	lastFinalizeTime := time.Now()
	lastLoadTime := time.Now()

	return &AuctionSessionManager{
		inMemoryAuctions,
		mutex,
		auctionRepo,
		false,
		lastAlertTime,
		lastFinalizeTime,
		lastLoadTime,
		alertCycle,
		finalizeCycle,
		loadAuctionCycle,
	}
}

func (auctionSessionManager *AuctionSessionManager) TurnOn() {
	if !auctionSessionManager.turnedOn {
		auctionSessionManager.turnedOn = true
		auctionSessionManager.lastAlertTime = time.Now()
		auctionSessionManager.lastFinalizeTime = time.Now()
		auctionSessionManager.lastLoadTime = time.Now()

		// bring into memory all auctions with lifecycles (start->end) overlapping with the following time span
		since := time.Now().Add(-loadBehindDuration) // assumes all auctions prior to this time have been finalized
		upTo := time.Now().Add(loadAheadDuration)
		auctionSessionManager.loadInAuctions(since, upTo)
		auctionSessionManager.loadInAuctions(since, upTo)
		auctionSessionManager.sendOutLifecycleAlerts()
		auctionSessionManager.finalizeAnyPastAuctions()

		go auctionSessionManager.intermittentlyLoadAuctions()
		go auctionSessionManager.intermittentlySendLifeCycleAlerts()
		go auctionSessionManager.intermittentlyFinalizeAuctions()
	}
}

func (auctionSessionManager *AuctionSessionManager) TurnOff() {
	if auctionSessionManager.turnedOn {
		auctionSessionManager.turnedOn = false // this will terminate 3 asynch goroutines
	}
}

func (auctionSessionManager *AuctionSessionManager) intermittentlyLoadAuctions() {
	for auctionSessionManager.turnedOn {
		if time.Since(auctionSessionManager.lastLoadTime) >= auctionSessionManager.loadCycle {
			since := auctionSessionManager.lastLoadTime.Add(loadAheadDuration)
			upTo := time.Now().Add(loadAheadDuration)
			auctionSessionManager.loadInAuctions(since, upTo) // acquires lock
			auctionSessionManager.lastLoadTime = time.Now()
		}
	}
}

func (auctionSessionManager *AuctionSessionManager) intermittentlySendLifeCycleAlerts() {
	for auctionSessionManager.turnedOn {
		if time.Since(auctionSessionManager.lastAlertTime) >= auctionSessionManager.alertCycle {
			// since := auctionSessionManager.lastAlertTime.Add(loadAheadDuration)
			// upTo := time.Now().Add(loadAheadDuration)
			auctionSessionManager.sendOutLifecycleAlerts() // acquires lock
			auctionSessionManager.lastAlertTime = time.Now()
		}
	}
}

func (auctionSessionManager *AuctionSessionManager) intermittentlyFinalizeAuctions() {
	for auctionSessionManager.turnedOn {
		if time.Since(auctionSessionManager.lastFinalizeTime) >= auctionSessionManager.finalizeCycle {
			// since := auctionSessionManager.lastLoadTime.Add(loadAheadDuration)
			// upTo := time.Now().Add(loadAheadDuration)
			auctionSessionManager.finalizeAnyPastAuctions() // acquires lock
			auctionSessionManager.lastFinalizeTime = time.Now()
		}
	}
}

func (auctionSessionManager *AuctionSessionManager) loadInAuctions(sinceTime time.Time, upToTime time.Time) {

	inMemAuctions := auctionSessionManager.inMemoryAuctions

	auctionSessionManager.mutex.Lock()

	auctions := auctionSessionManager.auctionRepository.GetAuctions(sinceTime, upToTime)
	for _, auction := range auctions {
		if !auction.IsFinalized() { // dont bring into memory if it is a finalized auction
			if _, ok := (*inMemAuctions)[auction.Item.ItemId]; !ok {
				(*inMemAuctions)[auction.Item.ItemId] = auction
			}
		}

	}
	auctionSessionManager.mutex.Unlock()
}

func (auctionSessionManager *AuctionSessionManager) sendOutLifecycleAlerts() {
	inMemAuctions := auctionSessionManager.inMemoryAuctions
	var sentNotif1, sentNotif2 bool

	auctionSessionManager.mutex.Lock()

	for _, auction := range *inMemAuctions {
		sentNotif1 = auction.SendStartSoonAlertIfApplicable()
		sentNotif2 = auction.SendEndSoonAlertIfApplicable()
		if sentNotif1 || sentNotif2 {
			auctionSessionManager.auctionRepository.SaveAuction(auction) // save the knowledge that alert was sent out;
		}
	}

	auctionSessionManager.mutex.Unlock()
}

func (auctionSessionManager *AuctionSessionManager) finalizeAnyPastAuctions() {
	inMemAuctions := auctionSessionManager.inMemoryAuctions

	auctionSessionManager.mutex.Lock()

	for _, auction := range *inMemAuctions {
		if time.Since(auction.Item.EndTime) >= FinalizeDelay && !auction.IsFinalized() {
			auction.Finalize()
			auctionSessionManager.auctionRepository.SaveAuction(auction) // save the knowledge that we finalized the auction
		}
	}

	auctionSessionManager.mutex.Unlock()
}
