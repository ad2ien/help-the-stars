package internal

import "time"

type HelpWantedIssue struct {
	Url              string
	Title            string
	IssueDescription string
	CreationDate     time.Time
	RepoOwner        string
	RepoDescription  string
	StargazersCount  int
}

type ThankStarsData struct {
	Issues            []HelpWantedIssue
	LastUpdate        time.Time
	CurrentlyUpdating bool
}
