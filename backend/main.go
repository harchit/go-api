package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func seedAccount(store Storage, firstname, lastname, pw string) *Account {
	//create new account object with all values
	acc, err := NewAccount(firstname, lastname, pw)
	if err != nil {
		log.Fatal(err)
	}

	//put into db
	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("new account => ", acc.Number)

	return acc
}

func seedAccounts(s Storage) {
	seedAccount(s, "dummy", "account", "32bhatoia")
	seedAccount(s, "dummy2", "account2", "32bhatoia")

}

func goDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

func main() {
	seed := flag.Bool("seed", false, "seed the database")
	flag.Parse()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	user := goDotEnvVariable("USERNAME")
	db := goDotEnvVariable("DB_NAME")
	pw := goDotEnvVariable("PASSWORD")
	ssl := goDotEnvVariable("SSL_MODE")

	store, err := NewPostgresStore(user, db, pw, ssl)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	if *seed {
		fmt.Println("seeding the database")
		seedAccounts(store)
	}
	server := NewApiServer(":3001", store)
	server.Run()

}
