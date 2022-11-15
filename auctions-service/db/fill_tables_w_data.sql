-- two calls below are basically saying "trunicate table if table exists"

CREATE TABLE IF NOT exists auctions (
    itemId varchar(255) PRIMARY KEY,
    sellerUserId varchar(255) NOT NULL,
	startPriceInCents NUMERIC(10, 2) NOT NULL,
    startTime timestamp(6) NOT NULL,
    endTime timestamp(6) NOT NULL
);
TRUNCATE TABLE auctions;

CREATE TABLE IF NOT exists auctionsCancellations (
    itemId varchar(255) PRIMARY KEY,
    timeCanceled timestamp(6) NOT NULL
);
TRUNCATE TABLE auctionscancellations;

CREATE TABLE IF NOT exists bids (
    bidId varchar(255) PRIMARY KEY,
    itemId varchar(255) NOT NULL,
    bidderUserId varchar(255) NOT NULL,
    amountInCents NUMERIC(10, 2) NOT NULL,
    timeBidProcessed timestamp(6) NOT NULL,
	active boolean NOT NULL
);
TRUNCATE TABLE bids;

-- insert some starter data

INSERT INTO auctions (itemId, sellerUserId, startPriceInCents, startTime, endTime) VALUES
	('200','270',600,TIMESTAMP '2022-12-01 15:00:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('201','336',5600,TIMESTAMP '2022-12-01 15:00:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('202','203',2100,TIMESTAMP '2022-12-01 15:00:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('203','247',4800,TIMESTAMP '2022-12-01 15:00:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('204','281',9700,TIMESTAMP '2022-12-01 15:00:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('205','248',1400,TIMESTAMP '2022-12-01 15:00:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('206','371',7500,TIMESTAMP '2022-12-01 15:00:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('207','215',3700,TIMESTAMP '2022-12-01 15:30:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('208','242',2200,TIMESTAMP '2022-12-01 15:30:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('209','226',3500,TIMESTAMP '2022-12-01 15:30:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('210','312',300,TIMESTAMP '2022-12-01 15:30:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('211','314',6300,TIMESTAMP '2022-12-01 15:30:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('212','292',4700,TIMESTAMP '2022-12-01 15:30:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('213','254',2200,TIMESTAMP '2022-12-01 15:30:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('214','301',6900,TIMESTAMP '2022-12-01 15:30:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('215','245',5500,TIMESTAMP '2022-12-01 15:30:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('216','256',6100,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('217','364',5600,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('218','256',11600,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('219','361',8700,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('220','256',2600,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('221','222',5600,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('222','269',2100,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('223','367',6400,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('224','241',8300,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('225','227',2200,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('226','225',7800,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000'),
	('227','263',3300,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 16:00:00.000000'),
	('228','323',400,TIMESTAMP '2022-12-01 15:45:00.000000',TIMESTAMP '2022-12-01 17:00:00.000000');


INSERT INTO auctionscancellations (itemId, timeCanceled) VALUES
	('209',TIMESTAMP '2022-11-01 16:00:00.000000'),
	('212',TIMESTAMP '2022-11-07 16:00:00.000000');

INSERT INTO bids (bidId, itemId, bidderUserId, amountInCents, timeBidProcessed, active) VALUES
	('100','200','431',623,TIMESTAMP '2022-12-01 15:15:00.000000',TRUE),
	('101','201','402',5613,TIMESTAMP '2022-12-01 15:16:00.000000',TRUE),
	('102','202','406',2111,TIMESTAMP '2022-12-01 15:15:00.000000',TRUE),
	('103','203','484',4830,TIMESTAMP '2022-12-01 15:18:00.000000',TRUE),
	('104','204','431',9702,TIMESTAMP '2022-12-01 15:19:00.000000',TRUE),
	('105','205','468',1424,TIMESTAMP '2022-12-01 15:16:00.000000',TRUE),
	('106','206','443',7514,TIMESTAMP '2022-12-01 15:18:00.000000',TRUE),
	('107','207','426',3725,TIMESTAMP '2022-12-01 15:31:00.000000',TRUE),
	('108','208','433',2216,TIMESTAMP '2022-12-01 15:32:00.000000',TRUE),
	-- ('109','209','404',3512,TIMESTAMP '2022-12-01 15:30:00.000000',TRUE),
	('110','210','435',322,TIMESTAMP '2022-12-01 15:40:00.000000',TRUE),
	('111','211','430',6315,TIMESTAMP '2022-12-01 15:41:00.000000',TRUE),
	-- ('112','212','480',4707,TIMESTAMP '2022-12-01 15:38:00.000000',TRUE),
	('113','213','440',2207,TIMESTAMP '2022-12-01 15:32:00.000000',TRUE),
	('114','214','427',6912,TIMESTAMP '2022-12-01 15:32:00.000000',TRUE),
	('115','215','482',5519,TIMESTAMP '2022-12-01 15:32:00.000000',TRUE),
	('116','216','413',6104,TIMESTAMP '2022-12-01 15:45:00.000000',TRUE),
	('117','217','496',5611,TIMESTAMP '2022-12-01 15:46:00.000000',TRUE),
	('118','218','438',11600,TIMESTAMP '2022-12-01 15:50:00.000000',TRUE),
	('119','219','418',8701,TIMESTAMP '2022-12-01 15:42:00.000000',TRUE),
	('120','220','429',2622,TIMESTAMP '2022-12-01 15:45:00.000000',TRUE),
	('121','221','478',5628,TIMESTAMP '2022-12-01 15:46:00.000000',TRUE),
	('122','222','499',2105,TIMESTAMP '2022-12-01 15:50:00.000000',TRUE),
	('123','223','406',6420,TIMESTAMP '2022-12-01 15:45:00.000000',TRUE),
	('124','224','488',8308,TIMESTAMP '2022-12-01 15:46:00.000000',TRUE),
	('125','225','427',2207,TIMESTAMP '2022-12-01 15:45:00.000000',TRUE),
	('126','226','451',7813,TIMESTAMP '2022-12-01 15:45:00.000000',TRUE),
	('127','227','459',3305,TIMESTAMP '2022-12-01 15:46:00.000000',TRUE),
	('128','228','462',420,TIMESTAMP '2022-12-01 15:50:00.000000',TRUE),
	('129','229','475',9500,TIMESTAMP '2022-12-01 15:42:00.000000',TRUE),
	('130','200','431',630,TIMESTAMP '2022-12-01 15:17:00.000000',TRUE),
	('131','201','463',5623,TIMESTAMP '2022-12-01 15:18:00.000000',TRUE),
	('132','202','496',2118,TIMESTAMP '2022-12-01 15:17:00.000000',TRUE),
	('133','203','420',4843,TIMESTAMP '2022-12-01 15:20:00.000000',TRUE),
	('134','204','409',9722,TIMESTAMP '2022-12-01 15:21:00.000000',TRUE),
	('135','205','446',1435,TIMESTAMP '2022-12-01 15:18:00.000000',TRUE),
	('136','206','447',7515,TIMESTAMP '2022-12-01 15:20:00.000000',TRUE),
	('137','207','454',3740,TIMESTAMP '2022-12-01 15:33:00.000000',FALSE),
	('138','208','472',2240,TIMESTAMP '2022-12-01 15:34:00.000000',TRUE),
	-- ('139','209','490',3528,TIMESTAMP '2022-12-01 15:32:00.000000',TRUE),
	('140','210','469',328,TIMESTAMP '2022-12-01 15:42:00.000000',TRUE),
	('141','211','493',6319,TIMESTAMP '2022-12-01 15:43:00.000000',TRUE),
	-- ('142','212','419',4717,TIMESTAMP '2022-12-01 15:40:00.000000',TRUE),
	('143','213','488',2218,TIMESTAMP '2022-12-01 15:34:00.000000',TRUE),
	('144','214','464',6934,TIMESTAMP '2022-12-01 15:34:00.000000',TRUE),
	('145','215','471',5526,TIMESTAMP '2022-12-01 15:34:00.000000',TRUE),
	('146','216','405',6113,TIMESTAMP '2022-12-01 15:47:00.000000',TRUE),
	('147','217','410',5633,TIMESTAMP '2022-12-01 15:48:00.000000',TRUE),
	('148','218','430',11603,TIMESTAMP '2022-12-01 15:52:00.000000',TRUE),
	('149','219','489',8713,TIMESTAMP '2022-12-01 15:44:00.000000',TRUE),
	('150','220','499',2627,TIMESTAMP '2022-12-01 15:47:00.000000',TRUE),
	('151','221','407',5658,TIMESTAMP '2022-12-01 15:48:00.000000',TRUE),
	('152','222','406',2122,TIMESTAMP '2022-12-01 15:52:00.000000',TRUE),
	('153','223','406',6439,TIMESTAMP '2022-12-01 15:47:00.000000',TRUE),
	('154','224','459',8322,TIMESTAMP '2022-12-01 15:48:00.000000',TRUE),
	('155','225','485',2218,TIMESTAMP '2022-12-01 15:47:00.000000',TRUE),
	('156','226','438',7815,TIMESTAMP '2022-12-01 15:47:00.000000',TRUE),
	('157','227','500',3327,TIMESTAMP '2022-12-01 15:48:00.000000',TRUE),
	('158','228','436',431,TIMESTAMP '2022-12-01 15:52:00.000000',TRUE),
	('159','229','486',9507,TIMESTAMP '2022-12-01 15:44:00.000000',TRUE);


select * from auctions a;
select * from auctionscancellations aC;
select * from bids b;
