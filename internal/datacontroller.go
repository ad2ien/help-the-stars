package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"help-the-stars/internal/persistence"
)

type DataController struct {
	queries *persistence.Queries
}

func CreateControler(db *sql.DB) *DataController {
	return &DataController{
		queries: persistence.New(db),
	}
}

func (d *DataController) GetAndSaveIssues() {

	// when, where from...
	fmt.Println("Loading issues...")
	data, err := GetStaredRepos(50)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	for i := 0; i < len(data); i++ {

		fmt.Println("Save an issue ", data[i].Url)
		d.queries.CreateIssue(ctx,
			mapModelToDbParameter(data[i]))
	}
}
