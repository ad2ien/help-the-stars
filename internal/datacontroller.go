package internal

import (
	"fmt"
	"log"
	"time"
)

type HelpWantedIssue struct {
	Title            string
	IssueDescription string
	URL              string
}

type HelpLookingRepo struct {
	RepoOwner       string
	RepoDescription string
	Issues          []HelpWantedIssue
}

type ThankStarsData struct {
	LastUpdate        time.Time
	HasNextPage       bool
	CurrentlyUpdating bool
	HelpLookingRepo   []HelpLookingRepo
}

func GetNextPage(index string) ThankStarsData {

	// when, where from...
	fmt.Println("Loading issues...")
	data, err := GetStaredRepos(50)
	if err != nil {
		log.Fatal(err)
	}

	return ThankStarsData{
		LastUpdate:        time.Now(),
		HasNextPage:       true,
		CurrentlyUpdating: false,
		HelpLookingRepo:   data,
	}
}
