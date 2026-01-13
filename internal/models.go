package internal

import "time"

type HelpWantedIssue struct {
	Url              string
	Title            string
	IssueDescription string
	CreationDate     time.Time
}

type Repo struct {
	RepoOwner             string
	RepoDescription       string
	StargazersCount       int
	LastIssueCreationTime time.Time
	Issues                []HelpWantedIssue
}

type ThankStarsData struct {
	Repos             []Repo
	LastUpdate        time.Time
	CurrentlyUpdating bool
}
