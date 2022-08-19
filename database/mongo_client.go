package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

func InitMongoDbClient() (*mongo.Client, func()) {
	URI := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(URI))
	if err != nil {
		panic(err)
	}

	return client, func() {
		client.Disconnect(context.Background())
	}
}

func InsertMany(db *mongo.Client, collectionName string, data []interface{}) {
	collection := db.Database("karting").Collection(collectionName)
	res, err := collection.InsertMany(context.Background(), data)
	if err != nil {
		log.Fatalln("Could not insert to mongo", err)
	} else {
		fmt.Println("Count Inserted:", len(res.InsertedIDs))
	}
}
