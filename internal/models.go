package internal

import "time"

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
