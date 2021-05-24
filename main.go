package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"context"
	"log"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"io/ioutil"
	"encoding/json"
	//"reflect"
	"strconv"
)

type Users struct{
	Email string `json:"email"`
	Role string `json:"role"`
	Status string `json:"status"`
}

type Items struct{
	//ID int `json:"id"`
	Name string `json:"name"`
	Room int32 `json:"room"`
	Quantity int32 `json:"quantity"`
	Remain int32 `json:"remain"`
}

type Log struct{
	//ID int `json:"id"`
	ID string `json:"id"`
	Action string `json:"action"`
	Email string `json:"email"`
	Item_Name string `json:"item_name"`
	BorrowTime string `json:"borrowtime"`
	ReturnTime string `json:"returntime"`
	Quantity int32 `json:"quantity"`
	Status string `json:"status"`
}

func testAPI(w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "rent-tool")
}

func getItem(w http.ResponseWriter, r *http.Request){

	var item Items
	var items []Items

	room := r.Header.Get("room")

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    itemsCollection := quickstartDatabase.Collection("items")

	intVar, _ := strconv.Atoi(room)

	filterCursor, err := itemsCollection.Find(ctx, bson.M{"room": intVar})
	if err != nil {
		log.Fatal(err)
	}
	var itemsFiltered []bson.M
	if err = filterCursor.All(ctx, &itemsFiltered); err != nil {
		log.Fatal(err)
	}

	for i, _ := range itemsFiltered {
		item.Name = itemsFiltered[i]["name"].(string)
		item.Room = itemsFiltered[i]["room"].(int32)
		item.Quantity = itemsFiltered[i]["quantity"].(int32)
		item.Remain = itemsFiltered[i]["remain"].(int32)

		items = append(items, item)
	}
	json.NewEncoder(w).Encode(items)
}

func postItem(w http.ResponseWriter, r *http.Request){

	type Input struct{
		Name string `json:"name"`
		Room int `json:"room"`
		Quantity int `json:"quantity"`
	}

	reqBody, _ := ioutil.ReadAll(r.Body)
	var input Input
	json.Unmarshal(reqBody, &input)

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    itemsCollection := quickstartDatabase.Collection("items")

	_, err = itemsCollection.InsertOne(ctx, bson.D{
		{Key: "name", Value: input.Name},
		{Key: "room", Value: input.Room},
		{Key: "quantity", Value: input.Quantity},
		{Key: "remain", Value: input.Quantity},
	})
	if err != nil {
		log.Fatal(err)
	}

}

func putItem(w http.ResponseWriter, r *http.Request){

	type Input struct{
		Name string `json:"name"`
		Room int `json:"room"`
		Quantity int `json:"quantity"`
	}

	reqBody, _ := ioutil.ReadAll(r.Body)
	var input Input
	json.Unmarshal(reqBody, &input)

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    itemsCollection := quickstartDatabase.Collection("items")

	_, err = itemsCollection.UpdateMany(
		ctx,
		bson.M{"name": input.Name},
		bson.D{
			{"$set", bson.D{{"room", input.Room}}},
			{"$set", bson.D{{"quantity", input.Quantity}}},
			{"$set", bson.D{{"remain", input.Quantity}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func transaction(w http.ResponseWriter, r *http.Request){
	type Input struct{
		Action string `json:"action"`
		ItemName string `json:"item_name"`
		Quantity int `json:"quantity"`
		Email string `json:"email"`
		ReturnTime string `json:"returntime"`
	}

	reqBody, _ := ioutil.ReadAll(r.Body)
	var input Input
	json.Unmarshal(reqBody, &input)

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    itemsCollection := quickstartDatabase.Collection("log")

	status := ""

	if input.Action == "borrow"{
		status = "complete"
	}else if input.Action == "return"{
		status = "pending"
	}

	_, err = itemsCollection.InsertOne(ctx, bson.D{
		{Key: "action", Value: input.Action},
		{Key: "item_name", Value: input.ItemName},
		{Key: "email", Value: input.Email},
		{Key: "quantity", Value: input.Quantity},
		{Key: "borrowtime", Value: time.Now().String()},
		{Key: "returntime", Value: input.ReturnTime},
		{Key: "quantity", Value: input.Quantity},
		{Key: "status", Value: status},
	})
	if err != nil {
		log.Fatal(err)
	}

	if input.Action == "borrow"{
		borrowItem(input.ItemName, int32(input.Quantity))
	}else if input.Action == "return"{
		returnItem(input.ItemName, int32(input.Quantity))
	}
}

func getUsers(w http.ResponseWriter, r *http.Request){

	var user Users
	var users []Users

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    usersCollection := quickstartDatabase.Collection("users")

	filterCursor, err := usersCollection.Find(ctx, bson.M{"role": "student"})
	if err != nil {
		log.Fatal(err)
	}
	var usersFiltered []bson.M
	if err = filterCursor.All(ctx, &usersFiltered); err != nil {
		log.Fatal(err)
	}

	for i, _ := range usersFiltered {
		user.Email = usersFiltered[i]["email"].(string)
		user.Status = usersFiltered[i]["status"].(string)
		user.Role = "student"

		users = append(users, user)
	}
	json.NewEncoder(w).Encode(users)
}

func postUsers(w http.ResponseWriter, r *http.Request){

	type Input struct{
		Email string `json:"email"`
		Role string `json:"role"`
		Status string `json:"status"`
	}

	reqBody, _ := ioutil.ReadAll(r.Body)
	var input Input
	json.Unmarshal(reqBody, &input)

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    usersCollection := quickstartDatabase.Collection("users")

	_, err = usersCollection.InsertOne(ctx, bson.D{
		{Key: "email", Value: input.Email},
		{Key: "role", Value: input.Role},
		{Key: "status", Value: input.Status},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func unbanUsers(w http.ResponseWriter, r *http.Request){

	type Input struct{
		Email string `json:"email"`
	}

	reqBody, _ := ioutil.ReadAll(r.Body)
	var input Input
	json.Unmarshal(reqBody, &input)

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    usersCollection := quickstartDatabase.Collection("users")

	_, err = usersCollection.UpdateMany(
		ctx,
		bson.M{"email": input.Email},
		bson.D{
			{"$set", bson.D{{"status", "normal"}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func banUsers(w http.ResponseWriter, r *http.Request){

	type Input struct{
		Email string `json:"email"`
	}

	reqBody, _ := ioutil.ReadAll(r.Body)
	var input Input
	json.Unmarshal(reqBody, &input)

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    usersCollection := quickstartDatabase.Collection("users")

	_, err = usersCollection.UpdateMany(
		ctx,
		bson.M{"email": input.Email},
		bson.D{
			{"$set", bson.D{{"status", "ban"}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func getLogs(w http.ResponseWriter, r *http.Request){

	var transaction Log
	var transactions []Log

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    logsCollection := quickstartDatabase.Collection("log")

	filterCursor, err := logsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	var logsFiltered []bson.M
	if err = filterCursor.All(ctx, &logsFiltered); err != nil {
		log.Fatal(err)
	}

	for i, _ := range logsFiltered {
		transaction.ID = logsFiltered[i]["_id"].(primitive.ObjectID).Hex()
		transaction.Action = logsFiltered[i]["action"].(string)
		transaction.Email = logsFiltered[i]["email"].(string)
		transaction.Item_Name = logsFiltered[i]["item_name"].(string)
		transaction.Quantity = logsFiltered[i]["quantity"].(int32)
		transaction.BorrowTime = logsFiltered[i]["borrowtime"].(string)
		transaction.ReturnTime = logsFiltered[i]["returntime"].(string)
		transaction.Status = logsFiltered[i]["status"].(string)
		transactions = append(transactions, transaction)
	}
	json.NewEncoder(w).Encode(transactions)
}

func changeTransactionStatus(w http.ResponseWriter, r *http.Request){

	type Input struct{
		ID string `json:"id"`
		Status string `json:"status"`
	}

	reqBody, _ := ioutil.ReadAll(r.Body)
	var input Input
	json.Unmarshal(reqBody, &input)

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    logsCollection := quickstartDatabase.Collection("log")

	id, _ := primitive.ObjectIDFromHex(input.ID)

	_, err = logsCollection.UpdateMany(
		ctx,
		bson.M{"_id": id},
		bson.D{
			{"$set", bson.D{{"status", input.Status}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func borrowItem(itemName string, quantity int32){
	var remainBefore int32
	var remainAfter int32

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    itemsCollection := quickstartDatabase.Collection("items")

	filterCursor, err := itemsCollection.Find(ctx, bson.M{"name": itemName})
	if err != nil {
		log.Fatal(err)
	}
	var itemsFiltered []bson.M
	if err = filterCursor.All(ctx, &itemsFiltered); err != nil {
		log.Fatal(err)
	}

	for i, _ := range itemsFiltered {
		remainBefore = itemsFiltered[i]["remain"].(int32)
	}

	remainAfter = remainBefore - quantity

	_, err = itemsCollection.UpdateMany(
		ctx,
		bson.M{"name": itemName},
		bson.D{
			{"$set", bson.D{{"remain", remainAfter}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func returnItem(itemName string, quantity int32){
	var remainBefore int32
	var remainAfter int32

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://206.189.156.219:27017"))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    quickstartDatabase := client.Database("Rent-Tool")
    itemsCollection := quickstartDatabase.Collection("items")

	filterCursor, err := itemsCollection.Find(ctx, bson.M{"name": itemName})
	if err != nil {
		log.Fatal(err)
	}
	var itemsFiltered []bson.M
	if err = filterCursor.All(ctx, &itemsFiltered); err != nil {
		log.Fatal(err)
	}

	for i, _ := range itemsFiltered {
		remainBefore = itemsFiltered[i]["remain"].(int32)
	}

	remainAfter = remainBefore + quantity

	_, err = itemsCollection.UpdateMany(
		ctx,
		bson.M{"name": itemName},
		bson.D{
			{"$set", bson.D{{"remain", remainAfter}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func handleRequests() {
	myRouter := mux.NewRouter()

	myRouter.HandleFunc("/",testAPI)
	myRouter.HandleFunc("/postitem",postItem).Methods("POST")
	myRouter.HandleFunc("/getitem",getItem).Methods("GET")
	myRouter.HandleFunc("/putitem",putItem).Methods("PUT")
	myRouter.HandleFunc("/transaction",transaction).Methods("POST")
	myRouter.HandleFunc("/getuser",getUsers).Methods("GET")
	myRouter.HandleFunc("/adduser",postUsers).Methods("POST")
	myRouter.HandleFunc("/unbanuser",unbanUsers).Methods("PUT")
	myRouter.HandleFunc("/banuser",banUsers).Methods("PUT")
	myRouter.HandleFunc("/getlogs",getLogs).Methods("GET")
	myRouter.HandleFunc("/changelogstatus",changeTransactionStatus).Methods("PUT")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET","POST","DELETE"},
		AllowCredentials: true,
		AllowedHeaders: []string{"*"},
	})

	http.ListenAndServe(":10000", c.Handler(myRouter))
}

func main() {
	handleRequests()
}