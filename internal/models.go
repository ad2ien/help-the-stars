package internal

import "time"

// Used to notify
type HelpWantedIssue struct {
	Url              string
	Title            string
	IssueDescription string
	CreationDate     time.Time
}

// Used to display an web interface
type Repo struct {
	RepoOwner             string
	RepoDescription       string
	StargazersCount       int
	Language              string
	LastIssueCreationTime time.Time
	Issues                []HelpWantedIssue
}

type ThankStarsData struct {
	Repos             []Repo
	LastUpdate        time.Time
	CurrentlyUpdating bool
}
