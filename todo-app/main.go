package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thedevsaddam/renderer"
	// New Official MongoDB Driver Imports
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive" // for ObjectId handling
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var rnd *renderer.Render
var client *mongo.Client // We use 'client' now, not 'db'

const (
	hostName       string = "localhost:27017"
	dbName         string = "demo_todo"
	collectionName string = "todo"
	port           string = ":9000"
)

type (
	// Use primitive.ObjectID instead of mgo/bson.ObjectId
	todoModel struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"`
		Title     string             `bson:"title"`
		Completed bool               `bson:"completed"`
		CreatedAt time.Time          `bson:"createAt"`
	}
	todo struct {
		ID        string    `json:"id,omitempty"`
		Title     string    `json:"title"`
		Completed bool      `json:"completed"`
		CreatedAt time.Time `json:"created_at,omitempty"`
	}
)

func init() {
	rnd = renderer.New()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect using the official driver
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+hostName))
	checkErr(err)
	log.Println("MongoDB Connected successfully")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
	checkErr(err)
}

func fetchTodos(w http.ResponseWriter, r *http.Request) {
	todos := []todoModel{}
	collection := client.Database(dbName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "failed to fetch todos",
			"error":   err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &todos); err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "failed to decode todos",
			"error":   err.Error(),
		})
		return
	}

	todoList := []todo{}
	for _, t := range todos {
		todoList = append(todoList, todo{
			ID:        t.ID.Hex(),
			Title:     t.Title,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt,
		})
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{"error": err.Error()})
		return
	}

	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title is required.",
		})
		return
	}

	tm := todoModel{
		ID:        primitive.NewObjectID(), // Use primitive package
		Title:     t.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	collection := client.Database(dbName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, tm)
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to save todo",
			"error":   err.Error(),
		})
		return
	}

	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "todo created successfully",
		"todo_id": tm.ID.Hex(),
	})
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(chi.URLParam(r, "id"))

	// Validate and convert the string ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	collection := client.Database(dbName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use DeleteOne with a filter for the specific ID
	res, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to delete todo",
			"error":   err.Error(),
		})
		return
	}

	if res.DeletedCount == 0 {
		rnd.JSON(w, http.StatusNotFound, renderer.M{
			"message": "Todo not found",
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "todo deleted successfully",
	})
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(chi.URLParam(r, "id"))

	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{"error": err.Error()})
		return
	}

	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "the title field is required",
		})
		return
	}

	collection := client.Database(dbName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Define the update operation using $set
	update := bson.M{"$set": bson.M{"title": t.Title, "completed": t.Completed}}

	// Use UpdateOne with a filter for the specific ID and the update document
	res, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)

	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to update todo",
			"error":   err.Error(),
		})
		return
	}
	
	if res.MatchedCount == 0 {
		rnd.JSON(w, http.StatusNotFound, renderer.M{
			"message": "Todo not found for update",
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "todo updated successfully",
	})
}


func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, os.Kill) // Explicitly handle interrupt/kill
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandlers())
	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Listening on port", port)
		// Check for specific error when server shuts down normally (http.ErrServerClosed)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	<-stopChan // Wait for OS interrupt signal

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Disconnect from MongoDB client before shutting down server
	if err := client.Disconnect(ctx); err != nil {
		log.Fatal(err)
	}

	srv.Shutdown(ctx)
	log.Println("Server and MongoDB connection gracefully stopped")
}

func todoHandlers() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}

// checkErr function specifically handles fatal errors during startup/init
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
