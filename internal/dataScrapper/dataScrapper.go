package dataScrapper

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type GameRelease struct {
	ID int64
	Title string
	Date time.Time
	platforms []string
	HasReleaseDate bool
	EstimatedRelease string
}

func scrapeGameData(url string, baseUrl string) ([]GameRelease) {
	gameslist := []GameRelease{}

	c := colly.NewCollector(
		colly.AllowedDomains(baseUrl),
	)
	c.OnHTML("div.c-entry-content > ul > li", func(e *colly.HTMLElement) {
	newGameRelease := GameRelease{}
	badTime := time.Time{}
			name := strings.Split(e.Text, "(")
			newGameRelease.Title = strings.TrimSpace(name[0])
			platforms := strings.Split(name[1], ")")
			platformsSlice := strings.Split(platforms[0],",")
			gameDate, estimatedReleaseStr, _ := getReleaseDate(e.Text)
			if gameDate.Equal(badTime) {
				newGameRelease.HasReleaseDate = false
				newGameRelease.EstimatedRelease = estimatedReleaseStr
			} else {
				newGameRelease.HasReleaseDate = true
				newGameRelease.Date = gameDate
				newGameRelease.EstimatedRelease = ""
			}
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

func getReleaseDate(str string) (time.Time, string, error) {
	parts := strings.Split(str, "—")
	
	var dateStr string
	if len(parts) > 1 {
		re := regexp.MustCompile(`\d`)
		output := re.ReplaceAllString(parts[1], "")
		fmt.Println(output)
		dateStr = strings.TrimSpace(parts[1])
	} else {
		return time.Time{}, "", fmt.Errorf("no time found")
	}
	dateParts := strings.Split(dateStr, " (") // Replace " - " with the character or string that marks the end of your date
    dateStr = dateParts[0]
	yearAddedDateStr := fmt.Sprintf("%s, 2024", dateStr)
	var layout string
    if strings.Contains(yearAddedDateStr, ".") && !strings.Contains(yearAddedDateStr, "Sept.") {
        layout = "Jan. 2, 2006"
    } else if strings.Contains(yearAddedDateStr, "Sept.") {
		date := strings.Split(yearAddedDateStr, ".")
		yearAddedDateStr = fmt.Sprintf("Sep. %s", date[1])
		layout = "Jan. 2, 2006"
	}  else if len(dateStr) > 3 && dateStr[3] == ' ' {
        layout = "Jan 2, 2006"
    }else {
        layout = "January 2, 2006"
    }
    date, err := time.Parse(layout, yearAddedDateStr)
	estimatedReleaseStr := yearAddedDateStr
    if err != nil {
        fmt.Println("Error:", err)
        return time.Time{}, estimatedReleaseStr, err
    }
    fmt.Println("Date:", date)
	return date, "", nil
}

func writeGameToDB(game GameRelease, db *sql.DB) (int64, error) {
	var id int64
	err := db.QueryRow(`INSERT INTO games(Title, release_date, hasreleasedate, estimated_released) VALUES($1, $2, $3, $4) RETURNING id`, game.Title, game.Date, game.HasReleaseDate, game.EstimatedRelease).Scan(&id)
	if err != nil {
        return 0, fmt.Errorf("writeGameToDB: %v", err)
    }
    if err != nil {
        return 0, fmt.Errorf("writeGameToDB: %v", err)
    }
    return id, nil
}

func RunDataScrape(db *sql.DB) (int, error) {
	SCRAPE_URL := os.Getenv("SCRAPE_URL")
	BASE_URL := os.Getenv("BASE_URL")
	gamesToBeAdded := scrapeGameData(SCRAPE_URL, BASE_URL)

	for i := 0; i < len(gamesToBeAdded); i++ {
		_, err := writeGameToDB(gamesToBeAdded[i], db)
		if err != nil {
			log.Fatal("Failed to save: ", gamesToBeAdded[i].Title)
		}
	}

	return len(gamesToBeAdded), nil

}