package main

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"

	"github.com/gin-gonic/gin"
	"net/http"
)

type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var collection *mongo.Collection

func main() {
	uri := "mongodb://localhost:27017"
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	collection = client.Database("blogging-api").Collection("albums")

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8080")
}

func getAlbums(c *gin.Context) {
	cursor, err := collection.Find(context.TODO(), map[string]interface{}{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	var albums []album

	for cursor.Next(context.TODO()) {
		var alb album
		if err := cursor.Decode(&alb); err != nil {
			log.Fatal(err)
		}
		albums = append(albums, alb)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, albums)
}

func postAlbums(c *gin.Context) {
	var newAlbum album

	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	filter := map[string]interface{}{"id": newAlbum.ID}

	var result album
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": "id already exists"})
		return
	}

	_, err = collection.InsertOne(context.TODO(), newAlbum)
	if err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	filter := map[string]interface{}{"id": id}

	var result album

	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
	}

	c.IndentedJSON(http.StatusOK, result)
}
