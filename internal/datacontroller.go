package internal

import (
	"context"
	"database/sql"
	"fmt"
	"help-the-stars/internal/persistence"
	"log"
	"time"
)

type DataController struct {
	queries *persistence.Queries
}

func CreateControler(db *sql.DB) *DataController {
	return &DataController{
		queries: persistence.New(db),
	}
}

func (d *DataController) GetAndSaveIssues() ThankStarsData {

	// when, where from...
	fmt.Println("Loading issues...")
	data, err := GetStaredRepos(50)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	for i := 0; i < len(data); i++ {

		for j := 0; j < len(data[i].Issues); j++ {
			fmt.Println("Save an issue ", data[i].Issues[j].URL)
			d.queries.CreateIssue(ctx, persistence.CreateIssueParams{
				Link:         data[i].Issues[j].URL,
				Title:        sql.NullString{String: data[i].Issues[j].Title, Valid: true},
				Description:  sql.NullString{String: data[i].Issues[j].IssueDescription, Valid: true},
				Owner:        sql.NullString{String: data[i].RepoOwner, Valid: true},
				CreationDate: sql.NullTime{Time: time.Now(), Valid: true},
			})
		}
	}

	return ThankStarsData{
		LastUpdate:        time.Now(),
		HasNextPage:       true,
		CurrentlyUpdating: false,
		HelpLookingRepo:   data,
	}
}
