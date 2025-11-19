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
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rnd *renderer.Renderer
var db *mgo.Database

const (
	hostName string = "localhost:27017"
	dbName string = "demo_todo"
	collectionName string = "todo"
	port string = ":9000"
)

type (
	todoModel struct{
		ID bson.ObjectId `bson:"_id,omitempty"`
		Title string `bson:"title"`
		Completed bool `bson:"completed"`
		CreatedAt time.Time `bson:"createAt"`
	}
	todo struct {
		ID string `json:"id"`
		Title string `json:"title"`
		Completed string `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}
)


func init(){
	rnd = renderer.New()
	sess, err:=mgo.Dial(hostname)
	checkErr(err)
	sess.SetMode(mgo.Monotonic, true)
	db = sess.DB(dbName)
}


func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOk, []string{"static/home.tpl"}, nil)
	checkErr(err)
}

func fetchTodos(w, http.ResponseWriter, r *http.Request){
	todos := []todoModel{}

	if err!=db.C(collectionName).Find(bson.M{}).All(&todos); err!=nil {
		rnd.JSON(w, http.StatusProcessing,renderer.M{
			"message":"failed to do todo",
			"error":err, 
		})
		return 
	}
	todoList := []todo{}

	for _,t := range todos{
		todoList = append(toList, todo{
			ID: t.ID.Hex(),
			Title: t.Title,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt
		})
	}
	rnd.JSON(w, http.StatusOk, renderer.M{
		"data": todoList,
	})

}

func main(){
	stopChan := make(chan os.signal)
	signal.Notify(stopChan)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/",homeHandler)
	r.Mount("/todo", todoHandlers())
	srv := &http.Server{
		Addr: port,
		Handler: r,
		ReadTimeout: 60*time.Second,
		WriteTimeout: 60*time.Second,
		IdleTimeout: 60*time.Second,
	}
	go func() {
		log.Println("Listening on port", port)
		if err:=srv.ListenAndServe(); err!=nil{
			log.Printf("listen:%s\n", err)
		}
	}()
	<-stopChan
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel(
		log.Println("server gracefully stopped")
	)
}


func todoHandlers() http.Handler{
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router){
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}

func checkErr(err){
	if err!=nil{
		log.Fatal(err)
	}
}