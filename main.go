package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

type GameRelease struct {
	ID int64
	title string
	date time.Time
	platforms []string
}


func main () {
	godotenv.Load()
	DB_NAME := os.Getenv("DB_NAME")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	SCRAPE_URL := os.Getenv("SCRAPE_URL")
	BASE_URL := os.Getenv("BASE_URL")
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

	gamesToBeAdded := scrapeGameData(SCRAPE_URL, BASE_URL)
	fmt.Println("New Game found: ", len(gamesToBeAdded))

	for i := 0; i < len(gamesToBeAdded); i++ {
		gameId, err := writeGameToDB(gamesToBeAdded[i])
		if err != nil {
			log.Fatal("Failed to save: ", gamesToBeAdded[i].title)
		}
		fmt.Println("Successfully added: ", gameId)
	}
}

func getReleaseDate (str string) (time.Time, error) {
	parts := strings.Split(str, "—")
	
	var dateStr string
	if len(parts) > 1 {
		re := regexp.MustCompile(`\d`)
		output := re.ReplaceAllString(parts[1], "")
		fmt.Println(output)
		dateStr = strings.TrimSpace(parts[1])
	} else {
		return time.Time{}, fmt.Errorf("no time found")
	}
	dateParts := strings.Split(dateStr, " (") // Replace " - " with the character or string that marks the end of your date
    dateStr = dateParts[0]
	yearAddedDateStr := fmt.Sprintf("%s, 2024", dateStr)
	var layout string
    if strings.Contains(yearAddedDateStr, ".") {
        layout = "Jan. 2, 2006"
    } else if len(dateStr) > 3 && dateStr[3] == ' ' {
        layout = "Jan 2, 2006"
    } else if strings.Contains(yearAddedDateStr, "Sept.") {
		layout = "Janua. 2, 2006"
	} else {
        layout = "January 2, 2006"
    }
    date, err := time.Parse(layout, yearAddedDateStr)
    if err != nil {
        fmt.Println("Error:", err)
        return time.Time{}, err
    }
    fmt.Println("Date:", date)
	return date, nil
}

func writeGameToDB (game GameRelease) (int64, error) {
	var id int64
	err := db.QueryRow(`INSERT INTO games(title, release_date) VALUES($1, $2) RETURNING id`, game.title, game.date).Scan(&id)
	if err != nil {
        return 0, fmt.Errorf("writeGameToDB: %v", err)
    }
    if err != nil {
        return 0, fmt.Errorf("writeGameToDB: %v", err)
    }
    return id, nil
}

func scrapeGameData (url string, baseUrl string) ([]GameRelease) {
	gameslist := []GameRelease{}

	c := colly.NewCollector(
		colly.AllowedDomains(baseUrl),
	)
	c.OnHTML("div.c-entry-content > ul > li", func(e *colly.HTMLElement) {
	newGameRelease := GameRelease{}
			name := strings.Split(e.Text, "(")
			newGameRelease.title = strings.TrimSpace(name[0])
			platforms := strings.Split(name[1], ")")
			platformsSlice := strings.Split(platforms[0],",")
			fmt.Println(platformsSlice)
			newGameRelease.date, _ = getReleaseDate(e.Text)
			for i := 0; i < len(platformsSlice); i++ {
				if len(platformsSlice) > 0 {
					newGameRelease.platforms = append(newGameRelease.platforms, platformsSlice[i])
				}
			}
			gameslist = append(gameslist, newGameRelease)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
        fmt.Println("Request URL: ", r.Request.URL, " failed with response: ", r, "\nError: ", err)
    })

	c.Visit(url)
	fmt.Println("Running...")
	return gameslist
}