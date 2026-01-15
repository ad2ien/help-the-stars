package internal

import "time"

type HasKey interface {
	Key() string
}

// Used to notify
type HelpWantedIssue struct {
	Url              string
	Title            string
	IssueDescription string
	CreationDate     time.Time
	RepoWithOwner    string
}

func (i HelpWantedIssue) Key() string {
	return i.Url
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

func (i Repo) Key() string {
	return i.RepoOwner
}

type ThankStarsData struct {
	Repos             []Repo
	LastUpdate        time.Time
	CurrentlyUpdating bool
}
