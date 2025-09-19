package main

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"net/http"
)

type post struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
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

	collection = client.Database("blogging-api").Collection("posts")

	router := gin.Default()
	router.GET("/posts", getAllPosts)
	router.GET("/posts/:id", getPostByID)
	router.POST("/posts", postPosts)
	router.PUT("/posts/:id", updatePost)
	router.DELETE("/posts/:id", deletePost)

	router.Run("localhost:8080")
}

func getAllPosts(c *gin.Context) {
	cursor, err := collection.Find(context.TODO(), map[string]interface{}{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	var posts []post

	for cursor.Next(context.TODO()) {
		var pst post
		if err := cursor.Decode(&pst); err != nil {
			log.Fatal(err)
		}
		posts = append(posts, pst)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, posts)
}

func postPosts(c *gin.Context) {
	var newPost post

	if err := c.BindJSON(&newPost); err != nil {
		return
	}

	filter := map[string]interface{}{"id": newPost.ID}

	var result post

	if err := collection.FindOne(context.TODO(), filter).Decode(&result); err == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": "id already exists"})
		return
	}

	if _, err := collection.InsertOne(context.TODO(), newPost); err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusCreated, newPost)
}

func getPostByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	filter := map[string]interface{}{"id": id}

	var result post

	if err := collection.FindOne(context.TODO(), filter).Decode(&result); err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, result)
}

func updatePost(c *gin.Context) {
	var update post

	id, _ := strconv.Atoi(c.Param("id"))

	filter := map[string]interface{}{"id": id}
	if err := c.BindJSON(&update); err != nil {
		return
	}

	updateQuery := map[string]interface{}{
		"$set": update,
	}

	result, err := collection.UpdateOne(context.TODO(), filter, updateQuery)
	if err != nil {
		log.Fatal(err)
	}

	if result.MatchedCount == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "post updated successfully"})
}

func deletePost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	filter := map[string]interface{}{"id": id}

	if _, err := collection.DeleteOne(context.TODO(), filter); err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "post delete successfully"})
}
