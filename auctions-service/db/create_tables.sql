DROP TABLE IF EXISTS auctions;
DROP TABLE IF EXISTS auctionsCancellations;
DROP TABLE IF EXISTS bids;

CREATE TABLE auctions (
    itemId varchar(255) PRIMARY KEY,
    sellerUserId varchar(255) NOT NULL,
    startPriceInCents NUMERIC(10, 2) NOT NULL,
    startTime timestamp(6) NOT NULL,
    endTime timestamp(6) NOT NULL
);

CREATE TABLE auctionsCancellations (
    itemId varchar(255) PRIMARY KEY,
    timeCanceled timestamp(6) NOT NULL
);

CREATE TABLE bids (
    bidId varchar(255) PRIMARY KEY,
    itemId varchar(255) NOT NULL,
    bidderUserId varchar(255) NOT NULL,
    amountInCents NUMERIC(10, 2) NOT NULL,
    timeBidProcessed timestamp(6) NOT NULL,
    active boolean NOT NULL
);
