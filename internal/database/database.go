package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/TyrinH/GameDayMap/internal/dataScrapper"
	"github.com/joho/godotenv"
)

func OpenDbConnection() (db *sql.DB) {
	godotenv.Load()
	DB_NAME := os.Getenv("DB_NAME")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	log.Print()
	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s", DB_USER, DB_NAME, DB_PASSWORD)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
	return db
}

func RetrieveAllGameReleases(db *sql.DB) ([]dataScrapper.GameRelease, error) {
	rows, queryErr := db.Query(`SELECT id, title, release_date, hasreleasedate, estimated_released FROM games;`)
	var gamesList []dataScrapper.GameRelease
	if queryErr != nil {
		log.Fatal(queryErr)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id int64
			title string
			release_date time.Time
			hasReleaseDate bool
			estimated_release string
		)
		err := rows.Scan(&id, &title, &release_date, &hasReleaseDate, &estimated_release)
		if err != nil {
			log.Fatal(err)
		}
		gamesList = append(gamesList, dataScrapper.GameRelease{ID: id, Title: title, Date: release_date, HasReleaseDate: hasReleaseDate, EstimatedRelease: estimated_release})
	}
	if err := rows.Err() 
	err != nil {
		return gamesList, err
	}
	return gamesList, nil
}

