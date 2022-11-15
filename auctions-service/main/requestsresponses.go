package main

import "auctions-service/domain"

type ResponseGetItemsByUserId struct {
	// could optionally include the userId here as well e.g.
	// "UserId string `json:"userid"`"
	ItemIds []string `json:"itemids"`
}

// type ResponseGetItemsByUserId struct {
// 	// could optionally include the userId here as well e.g.
// 	// "UserId string `json:"userid"`"
// 	ItemIds []string `json:"itemids"`
// }

type ResponseStopAuction struct {
	Msg string `json:"message"`
}

type RequestStopAuction struct {
	// ItemId            string `json:"itemid"`
	SellerUserId string `json:"selleruserid"`
}

type RequestCreateAuction struct {
	ItemId            string `json:"itemid"`
	SellerUserId      string `json:"selleruserid"`
	StartTime         string `json:"starttime"`
	EndTime           string `json:"endtime"`
	StartPriceInCents int64  `json:"startpriceincents"`
}

type ResponseCreateAuction struct {
	Msg string `json:"message"`
}

type ResponseGetActiveAuctions struct {
	ActiveAuctions []JsonAuction `json:"activeauctions"`
}

type JsonAuction struct {
	ItemId            string `json:"itemid"`
	SellerUserId      string `json:"selleruserid"`
	StartTime         string `json:"starttime"`
	EndTime           string `json:"endtime"`
	StartPriceInCents int64  `json:"startpriceincents"`
}

func ExportAuction(auction *domain.Auction) *JsonAuction {
	layout := "2006-01-02 15:04:05.000000"
	return &JsonAuction{
		ItemId:            auction.Item.ItemId,
		SellerUserId:      auction.Item.SellerUserId,
		StartPriceInCents: auction.Item.StartPriceInCents,
		StartTime:         auction.Item.StartTime.Format(layout),
		EndTime:           auction.Item.EndTime.Format(layout),
	}
}

// type ResponseCreateAuction struct {

// }

// Item         *Item
// bids         []*Bid // slice of pointers to bids; new higher bids get appended on the end
// cancellation *Cancellation
