package main

import (
	"auctions-service/domain"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"              // acquired by doing 'go get github.com/gorilla/mux.git'
	_ "github.com/lib/pq"                 // postgres
	amqp "github.com/rabbitmq/amqp091-go" // acquired by doing 'go get github.com/rabbitmq/amqp091-go'
)

var Articles []Article // like a database
var db *sql.DB

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

type Article struct {
	Title   string `json:"Title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

func getrowsindb(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var res CustomerData
		// var todos []string
		rows, err := db.Query("SELECT * FROM customer LIMIT 10")
		defer rows.Close()
		if err != nil {
			log.Fatalln(err)
			// c.JSON("An error occured")
		}
		for rows.Next() {
			if err := rows.Scan(
				&res.Customer_id,
				&res.Store_id,
				&res.First_name,
				&res.Last_name,
				&res.Email,
				&res.Address_id,
				&res.Activebool,
				&res.Create_date,
				&res.Last_update,
				&res.Active,
			); err != nil {
				fmt.Println(err.Error())
			}
			// todos = append(todos, res)
			fmt.Println(res)
		}
	}

}

func cancelAuction(auctionservice *AuctionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var res itemIds
		vars := mux.Vars(r)
		itemId := vars["itemId"]

		// reqBody, _ := ioutil.ReadAll(r.Body) // read details

		var requestBody RequestStopAuction // parse request into a struct with assumed structure
		// err := json.Unmarshal(reqBody, &requestBody)
		err := json.NewDecoder(r.Body).Decode(&requestBody)

		var response ResponseStopAuction

		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Msg = "request body was ill-formed"

			// json.Marshal()
			json.NewEncoder(w).Encode(response)
			return
			// w.Write(js)
		}

		// itemId := requestBody.SellerUserId
		requesterUserId := requestBody.SellerUserId
		cancelAuctionOutcome := auctionservice.CancelAuction(itemId, requesterUserId)

		if cancelAuctionOutcome == auctionNotExist {
			response.Msg = "auction does not exist."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if cancelAuctionOutcome == auctionCancellationRequesterIsNotSeller {
			response.Msg = "requesting user is not the seller of the item in auction. Not allowed to cancel auction."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if cancelAuctionOutcome == auctionAlreadyOver {
			response.Msg = "auction is already over."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if cancelAuctionOutcome == auctionAlreadyCanceled {
			response.Msg = "auction has already been canceled."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return

		}

		// success
		if cancelAuctionOutcome == auctionSuccessfullyCanceled {
			response.Msg = "successfully stopped auction."
			json.NewEncoder(w).Encode(response)
			return
		}

		panic("see cancelAuction() in main.go; could not determine an outcome for cancel Auction request")

	}
}

func getItemsUserHasBidsOn(auctionservice *AuctionService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// var res itemIds
		vars := mux.Vars(r)
		userId := vars["userId"]
		fmt.Println(userId)
		itemIds := auctionservice.GetItemsUserHasBidsOn(userId)
		fmt.Println(itemIds)

		response := ResponseGetItemsByUserId{*itemIds}

		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// json.NewEncoder(w).Encode(article)

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		// // var todos []string
		// rows, err := db.Query("SELECT * FROM customer LIMIT 10")
		// defer rows.Close()
		// if err != nil {
		// 	log.Fatalln(err)
		// 	// c.JSON("An error occured")
		// }
		// for rows.Next() {
		// 	if err := rows.Scan(
		// 		&res.Customer_id,
		// 		&res.Store_id,
		// 		&res.First_name,
		// 		&res.Last_name,
		// 		&res.Email,
		// 		&res.Address_id,
		// 		&res.Activebool,
		// 		&res.Create_date,
		// 		&res.Last_update,
		// 		&res.Active,
		// 	); err != nil {
		// 		fmt.Println(err.Error())
		// 	}
		// 	// todos = append(todos, res)
		// 	fmt.Println(res)
		// }
	}

}

func getActiveAuctions(auctionservice *AuctionService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// var res itemIds
		// vars := mux.Vars(r)
		// userId := vars["userId"]
		// fmt.Println(userId)
		activeAuctions := auctionservice.GetActiveAuctions()
		// fmt.Println(itemIds)
		exportedAuctions := make([]JsonAuction, len(*activeAuctions))
		for i, activeAuction := range *activeAuctions {
			exportedAuctions[i] = *ExportAuction(activeAuction)
		}

		response := ResponseGetActiveAuctions{exportedAuctions}

		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// json.NewEncoder(w).Encode(article)

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		// // var todos []string
		// rows, err := db.Query("SELECT * FROM customer LIMIT 10")
		// defer rows.Close()
		// if err != nil {
		// 	log.Fatalln(err)
		// 	// c.JSON("An error occured")
		// }
		// for rows.Next() {
		// 	if err := rows.Scan(
		// 		&res.Customer_id,
		// 		&res.Store_id,
		// 		&res.First_name,
		// 		&res.Last_name,
		// 		&res.Email,
		// 		&res.Address_id,
		// 		&res.Activebool,
		// 		&res.Create_date,
		// 		&res.Last_update,
		// 		&res.Active,
		// 	); err != nil {
		// 		fmt.Println(err.Error())
		// 	}
		// 	// todos = append(todos, res)
		// 	fmt.Println(res)
		// }
	}

}

func stopAuction(auctionservice *AuctionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var res itemIds
		vars := mux.Vars(r)
		itemId := vars["itemId"]

		// var requestBody RequestCreateAuction // parse request into a struct with assumed structure
		var response ResponseStopAuction

		w.Header().Set("Content-Type", "application/json")

		fmt.Println(itemId)
		stopAuctionOutcome := auctionservice.StopAuction(itemId)

		if stopAuctionOutcome == auctionNotExist {
			response.Msg = "auction does not exist."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if stopAuctionOutcome == auctionAlreadyCanceled {
			response.Msg = "auction has already been canceled."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return

		}

		if stopAuctionOutcome == auctionAlreadyOver {
			response.Msg = "auction is already over."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// success
		if stopAuctionOutcome == auctionSuccessfullyStopped {
			response.Msg = "successfully stopped auction."
			json.NewEncoder(w).Encode(response)
			return
		}

		panic("see stopAuction() in main.go; could not determine an outcome for stop Auction request")

	}
}

func createAuction(auctionservice *AuctionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var res itemIds
		// vars := mux.Vars(r)
		// itemId := vars["itemId"]

		// reqBody, _ := ioutil.ReadAll(r.Body) // read details

		var requestBody RequestCreateAuction // parse request into a struct with assumed structure
		// err := json.Unmarshal(reqBody, &requestBody)
		err := json.NewDecoder(r.Body).Decode(&requestBody)

		fmt.Println(requestBody)
		var response ResponseCreateAuction

		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Msg = "request body was ill-formed"

			// json.Marshal()
			json.NewEncoder(w).Encode(response)
			return
			// w.Write(js)
		}

		// interpret body

		// js, err := json.Marshal(response)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		itemId := requestBody.ItemId
		sellerUserId := requestBody.SellerUserId
		startTime, err1 := interpretTimeStr(requestBody.StartTime)
		endTime, err2 := interpretTimeStr(requestBody.EndTime)
		startPriceInCents := requestBody.StartPriceInCents

		if err1 != nil || err2 != nil {
			response.Msg = "startTime or endTime was not given in expected format: use YYYY-MM-DD HH:MM:SS.SSSSSS"
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		createAuctionOutcome := auctionservice.CreateAuction(itemId, sellerUserId, startTime, endTime, startPriceInCents)

		// var response ResponseCreateAuction

		if createAuctionOutcome == auctionAlreadyCreated {
			response.Msg = "an auction already exists for this item."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if createAuctionOutcome == auctionStartsInPast {
			response.Msg = "auction would start in the past."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if createAuctionOutcome == auctionWouldStartTooSoon {
			response.Msg = "an auction cannot be created within 2 hours before auction start. schedule the auction for a later time."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if createAuctionOutcome == badTimeSpecified {
			response.Msg = "startTime is not < endTime."
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// success
		if createAuctionOutcome == auctionSuccessfullyCreated {
			response.Msg = "successfully created auction."
			// w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		panic("see createAuction() in main.go; could not determine an outcome for create Auction request")

		// // var todos []string
		// rows, err := db.Query("SELECT * FROM customer LIMIT 10")
		// defer rows.Close()
		// if err != nil {
		// 	log.Fatalln(err)
		// 	// c.JSON("An error occured")
		// }
		// for rows.Next() {
		// 	if err := rows.Scan(
		// 		&res.Customer_id,
		// 		&res.Store_id,
		// 		&res.First_name,
		// 		&res.Last_name,
		// 		&res.Email,
		// 		&res.Address_id,
		// 		&res.Activebool,
		// 		&res.Create_date,
		// 		&res.Last_update,
		// 		&res.Active,
		// 	); err != nil {
		// 		fmt.Println(err.Error())
		// 	}
		// 	// todos = append(todos, res)
		// 	fmt.Println(res)
		// }
	}
}

func processNewBid(auctionservice *AuctionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var res itemIds
		// vars := mux.Vars(r)
		// itemId := vars["itemId"]

		// reqBody, _ := ioutil.ReadAll(r.Body) // read details

		var requestBody RequestProcessNewBid // parse request into a struct with assumed structure
		// err := json.Unmarshal(reqBody, &requestBody)
		err := json.NewDecoder(r.Body).Decode(&requestBody)

		fmt.Println(requestBody)
		var response ResponseProcessNewBid

		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Msg = "request body was ill-formed"

			// json.Marshal()
			json.NewEncoder(w).Encode(response)
			return
			// w.Write(js)
		}

		// interpret body

		// js, err := json.Marshal(response)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		itemId := requestBody.ItemId
		bidderUserId := requestBody.BidderUserId
		timeReceived := time.Now()
		amountInCents := requestBody.AmountInCents

		if amountInCents < 0 {
			response.Msg = "bid money amount was negative integer."
			response.WasNewTopBid = false
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		auctionInteractionOutcome, auctionState, wasNewTopBid := auctionservice.ProcessNewBid(itemId, bidderUserId, timeReceived, amountInCents)

		// createAuctionOutcome := auctionservice.CreateAuction(itemId, sellerUserId, startTime, endTime, startPriceInCents)

		// var response ResponseCreateAuction

		if auctionInteractionOutcome == auctionNotExist {
			response.Msg = "auction does not exist."
			response.WasNewTopBid = false
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if auctionState == domain.PENDING {
			response.Msg = "auction has not yet started."
			response.WasNewTopBid = false
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if auctionState == domain.OVER {
			response.Msg = "auction is already over."
			response.WasNewTopBid = false
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if auctionState == domain.FINALIZED {
			response.Msg = "auction has already been finalized (archived)."
			response.WasNewTopBid = false
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// success case 1
		if auctionState == domain.ACTIVE && !wasNewTopBid {
			response.Msg = "bid was not a new top bid because it was under start price or under the current top bid price."
			response.WasNewTopBid = false
			json.NewEncoder(w).Encode(response)
			return
		}

		// success case 1
		if auctionState == domain.ACTIVE && wasNewTopBid {
			response.Msg = "successfully processed bid; bid was new top bid!"
			response.WasNewTopBid = true
			json.NewEncoder(w).Encode(response)
			return
		}

		panic("see processNewBid() in main.go; could not determine an outcome for place new Bid request")

	}
}

func interpretTimeStr(timeStr string) (*time.Time, error) {
	layout := "2006-01-02 15:04:05.000000"
	t, err := time.Parse(layout, timeStr)
	return &t, err
}

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllArticles")
	json.NewEncoder(w).Encode(Articles)
}

func returnItemsUserHasBidsOn(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllArticles")
	json.NewEncoder(w).Encode(Articles)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func publishNotif(w http.ResponseWriter, r *http.Request) {
	// make connection
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq-server:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// declare queue for us to send messages to
	q, err := ch.QueueDeclare(
		"notifications", // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := "Hello World!"
	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", body)
}

// method that when executed spawns a goroutine to listen for incoming
// messages on a queue for new bids. With each new bid that appears
// in the queue, this method calls upon the auctionservice to process
// the new bid
func handleNewBids(auctionservice *AuctionService) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq-server:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"notifications", // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			// characterize
			// auctionservice.ProcessNewBid()
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func handleHTTPAPIRequests(auctionservice *AuctionService) {
	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	// myRouter.HandleFunc("/", homePage)
	// myRouter.HandleFunc("/all", returnAllArticles)
	// myRouter.HandleFunc("/publishNotifc", publishNotif)
	// myRouter.HandleFunc("/rowsindb", getrowsindb(db))
	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument

	// define all REST/HTTP API endpoints below
	apiVersion := "v1"
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc(fmt.Sprintf("/api/%s/Auctions/", apiVersion), createAuction(auctionservice)).Methods("POST")
	myRouter.HandleFunc(fmt.Sprintf("/api/%s/Bids/", apiVersion), processNewBid(auctionservice)).Methods("POST")
	myRouter.HandleFunc(fmt.Sprintf("/api/%s/cancelAuction/{itemId}", apiVersion), cancelAuction(auctionservice))
	myRouter.HandleFunc(fmt.Sprintf("/api/%s/stopAuction/{itemId}", apiVersion), stopAuction(auctionservice))
	myRouter.HandleFunc(fmt.Sprintf("/api/%s/ItemsUserHasBidsOn/{userId}", apiVersion), getItemsUserHasBidsOn(auctionservice)).Methods("GET")
	myRouter.HandleFunc(fmt.Sprintf("/api/%s/activeAuctions/", apiVersion), getActiveAuctions(auctionservice)).Methods("GET")
	// get active auctions

	// myRouter.HandleFunc("/publishNotifc", publishNotif)

	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

// func handleRabbitMQEvents() {

// }

func main() {
	fmt.Println("Rest API v2.0 - Mux Routers")
	// Connect to database
	// connStr := "postgresql://postgres:mysecret@localhost:5432/dvdrental?sslmode=disable"
	// db, err := sql.Open("postgres", connStr)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// go handleHTTPAPIRequests(db)
	// Articles = []Article{
	// 	Article{Title: "Hello", Desc: "Article Description", Content: "Article Content"},
	// 	Article{Title: "Hello 2", Desc: "Article Description", Content: "Article Content"},
	// }
	bidRepo := domain.NewInMemoryBidRepository(false) // do not use seed; assign random uuid's
	auctionRepo := domain.NewInMemoryAuctionRepository()

	// fill bid repo with some bids
	time1 := time.Date(2014, 2, 4, 00, 00, 00, 0, time.UTC)
	time2 := time.Date(2014, 2, 4, 00, 00, 00, 0, time.UTC)    // same as time1
	time3 := time.Date(2014, 2, 4, 00, 00, 00, 1000, time.UTC) // 1 microsecond after
	time4 := time.Date(2014, 2, 4, 00, 00, 01, 0, time.UTC)    // 1 sec after
	bid1 := *domain.NewBid("101", "20", "asclark109", time1, int64(300), true)
	bid2 := *domain.NewBid("102", "20", "mcostigan9", time2, int64(300), true)
	bid3 := *domain.NewBid("103", "20", "katharine2", time3, int64(400), true)
	bid4 := *domain.NewBid("104", "20", "katharine2", time4, int64(10), true)
	bidRepo.SaveBid(&bid1)
	bidRepo.SaveBid(&bid2)
	bidRepo.SaveBid(&bid3)
	bidRepo.SaveBid(&bid4)

	startime := time.Date(2014, 2, 4, 01, 00, 00, 0, time.UTC)
	endtime := time.Date(2014, 2, 4, 01, 30, 00, 0, time.UTC)                    // 30 min later
	item1 := domain.NewItem("101", "asclark109", startime, endtime, int64(2000)) // $20 start price
	item2 := domain.NewItem("102", "asclark109", startime, endtime, int64(2000)) // $20 start price
	auction1 := domain.NewAuction(item1, nil, nil, false, false, nil)            // will go to completion
	auction2 := domain.NewAuction(item2, nil, nil, false, false, nil)            // will get cancelled halfway through

	nowtime := time.Now()
	latertime := nowtime.Add(time.Duration(4) * time.Hour)                        // 4 hrs from now
	item3 := domain.NewItem("103", "asclark109", nowtime, latertime, int64(2000)) // $20 start price
	auctionactive := domain.NewAuction(item3, nil, nil, false, false, nil)

	latertime2 := nowtime.Add(time.Duration(2) * time.Hour)                        // 2 hrs from now
	item4 := domain.NewItem("104", "asclark109", nowtime, latertime2, int64(2000)) // $20 start price
	auctionactive2 := domain.NewAuction(item4, nil, nil, false, false, nil)

	auctionRepo.SaveAuction(auction1)
	auctionRepo.SaveAuction(auction2)
	auctionRepo.SaveAuction(auctionactive)
	auctionRepo.SaveAuction(auctionactive2)

	auctionservice := NewAuctionService(bidRepo, auctionRepo)

	alertCycle := time.Duration(1) * time.Minute
	finalizeCycle := time.Duration(1) * time.Minute
	loadAuctionCycle := time.Duration(1) * time.Minute
	auctionSessionManager := NewAuctionSessionManager(auctionservice, alertCycle, finalizeCycle, loadAuctionCycle)
	auctionSessionManager.TurnOn()

	go handleHTTPAPIRequests(auctionservice)
	go handleNewBids(auctionservice)

	var forever chan struct{}
	<-forever
}

type CustomerData struct {
	Customer_id string
	Store_id    int64
	First_name  string
	Last_name   string
	Email       string
	Address_id  string
	Activebool  string
	Create_date string
	Last_update string
	Active      string
}
