package main

import (
	"context"
	"encoding/json"
	"log"

	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	client *mongo.Client
}

func NewServer(c *mongo.Client) *Server {
	return &Server{
		client: c,
	}
}

func (s *Server) handleGetAllFacts(w http.ResponseWriter, r *http.Request) {
	collection := s.client.Database("catFact").Collection("facts")

	query := bson.M{}
	cursor, err := collection.Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}

	results := []bson.M{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

type CatFactWorker struct {
	client *mongo.Client
}

func NewCatFactWorker(c *mongo.Client) *CatFactWorker {
	return &CatFactWorker{
		client: c,
	}
}

func (cfw *CatFactWorker) start() error {
	collection := cfw.client.Database("catFact").Collection("facts")
	ticker := time.NewTicker(2 * time.Second)

	for {
		resp, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			return err
		}
		var catFact bson.M
		if err := json.NewDecoder(resp.Body).Decode(&catFact); err != nil {
			return err
		}

		_, err = collection.InsertOne(context.TODO(), catFact)
		if err != nil {
			return err
		}

		<-ticker.C
	}
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	server := NewServer(client)

	worker := NewCatFactWorker(client)
	go worker.start()

	http.HandleFunc("/facts", server.handleGetAllFacts)
	http.ListenAndServe(":3000", nil)
}
