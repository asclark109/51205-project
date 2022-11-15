package domain

type inMemoryBidRepository struct {
	bids []*Bid
}

func NewInMemoryBidRepository() BidRepository {
	bids := []*Bid{}
	return &inMemoryBidRepository{bids}
}

func (repo *inMemoryBidRepository) GetBid(bidId string) *Bid {
	for _, bid := range repo.bids {
		if bid.BidId == bidId {
			return bid
		}
	}
	return nil
}

func (repo *inMemoryBidRepository) GetBidsByUserId(biddeUserId string) *[]*Bid {
	relevantBids := []*Bid{}
	for _, bid := range repo.bids {
		if bid.BidderUserId == biddeUserId {
			relevantBids = append(relevantBids, bid)
		}
	}
	return &relevantBids
}

func (repo *inMemoryBidRepository) GetBidsByItemId(itemId string) *[]*Bid {
	relevantBids := []*Bid{}
	for _, bid := range repo.bids {
		if bid.ItemId == itemId {
			relevantBids = append(relevantBids, bid)
		}
	}
	return &relevantBids
}

func (repo *inMemoryBidRepository) SaveBid(bidToSave *Bid) {
	for idx, bid := range repo.bids {
		if bid.BidId == bidToSave.BidId {
			repo.bids[idx] = bidToSave // overwrite
		}
	}
	// else its new
	repo.bids = append(repo.bids, bidToSave)
}
