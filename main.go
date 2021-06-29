package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	person = &Person{
		Name: "Jack", Email: "jack@email.com",
	}
	books = []Book{
		{Title: "The Rules fo Thinking", Author: "Richard Templer", CallNumber: 1234, PersonID: 1},
		{Title: "Book 2", Author: "Author 2", CallNumber: 2345, PersonID: 1},
		{Title: "Book 3", Author: "Author 3", CallNumber: 3456, PersonID: 1},
	}
)

var db *gorm.DB
var err error

type Person struct {
	gorm.Model // Makes sure a Person has ID

	Name  string
	Email string `gorm:"type varchar(100); unique_index"`
	Books []Book
}
type Book struct {
	gorm.Model

	Title      string
	Author     string
	CallNumber int `gorm:"unique_index"`
	PersonID   int
	//Foreign key relationship : Have to use
	//the name of the struct+ID
}

func main() {
	//Loading environment variables for DATABASE connection
	dialect := os.Getenv("DIALECT")
	host := os.Getenv("HOST")
	dbPort := os.Getenv("DBPORT")
	user := os.Getenv("USER")
	dbName := os.Getenv("NAME")
	password := os.Getenv("PASSWORD")

	// Database connection string
	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s port=%s", host, user, dbName, password, dbPort)

	// Opening connection to database
	db, err = gorm.Open(dialect, dbURI)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Successfully connected to database")
	}

	// Close connection to database when the main function finishes
	defer db.Close()

	// Make migration to the database if they have not been already created
	db.AutoMigrate(&Person{})
	db.AutoMigrate(&Book{})

	// API routes
	router := mux.NewRouter()

	router.HandleFunc("/people", getPeople).Methods("GET")
	router.HandleFunc("/person/{id}", getPerson).Methods("GET") // get a person and all his books
	router.HandleFunc("/books", getBooks).Methods("GET")        // read all books from databse
	router.HandleFunc("/book/{id}", getBook).Methods("GET")

	router.HandleFunc("/create/person", createPerson).Methods("POST")
	router.HandleFunc("/create/book", createBook).Methods("POST")

	router.HandleFunc("/delete/person/{id}", deletePerson).Methods("DELETE")
	router.HandleFunc("/delete/book/{id}", deleteBook).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))
}

// API Controllers

// controller of Persons
func getPeople(w http.ResponseWriter, r *http.Request) {
	var people []Person

	db.Find(&people)

	json.NewEncoder(w).Encode(&people)
}

func getPerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var person Person
	var books []Book

	// find the first match from database
	db.First(&person, params["id"])
	db.Model(&person).Related(&books)

	person.Books = books

	json.NewEncoder(w).Encode(person)
}

// Somebody will send person data as JSON
// and we will put it into person struct and then into database
func createPerson(w http.ResponseWriter, r *http.Request) {
	var person Person
	json.NewDecoder(r.Body).Decode(&person)

	createdPerson := db.Create(&person)
	err = createdPerson.Error
	if err != nil {
		// send the error to the URL endpoint instead of created person in case of error
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(&person)
	}
}

func deletePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var person Person

	db.First(&person, params["id"])
	db.Delete(&person)

	json.NewEncoder(w).Encode(&person)
}

// Book Controllers

func getBooks(w http.ResponseWriter, r *http.Request) {
	var books []Book

	db.Find(&books)

	json.NewEncoder(w).Encode(&books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r) //grab the ID of the book fromt he URL as JSON(i think)

	var book Book

	db.First(&book, params["id"])

	json.NewEncoder(w).Encode(&book)

}

func createBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	json.NewDecoder(r.Body).Decode(&book)

	createdBook := db.Create(&book)
	err = createdBook.Error
	if err != nil {
		// send the error to the URL endpoint instead of created book in case of error
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(&book)
	}
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var book Book

	db.First(&book, params["id"])
	db.Delete(&book)

	json.NewEncoder(w).Encode(&book)
}
