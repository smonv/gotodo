package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbUser string = "postgres"
	dbPass string = "mashiro"
	dbName        = "gotodo"
)

var db *sql.DB

type Todo struct {
	Id          int
	Title       string
	Description string
	Created_at  time.Time
}

func main() {
	var err error
	dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPass, dbName)
	db, err = sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/", todosIndex)

	http.ListenAndServe(":8080", nil)
}

func todosIndex(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM todo")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	todos := make([]*Todo, 0)
	for rows.Next() {
		todo := new(Todo)
		err := rows.Scan(&todo.Id, &todo.Title, &todo.Description, &todo.Created_at)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	indexTemplate(w, r, todos)
}

func indexTemplate(w http.ResponseWriter, r *http.Request, todos []*Todo) {
	lp := path.Join("templates", "layout.html")
	fp := path.Join("templates", "index.html")

	// Return a 404 if the template doesn't exist
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
	}

	// Return a 404 if the request is for a directory
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		// Log the detailed error
		log.Println(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", todos); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}
