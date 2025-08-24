package internal

func mapGhQueryToHelpWantedIssue(query GhQuery) []HelpLookingRepo {
	var helpLookingRepos []HelpLookingRepo

	for _, repo := range query.Viewer.StarredRepositories.Nodes {
		if repo.Issues.Nodes == nil ||
			len(repo.Issues.Nodes) == 0 {
			continue
		}
		helpLookingRepo := HelpLookingRepo{
			RepoOwner:       string(repo.NameWithOwner),
			RepoDescription: string(repo.Description),
		}
		for _, issue := range repo.Issues.Nodes {
			helpWantedIssue := HelpWantedIssue{
				Title:            string(issue.Title),
				IssueDescription: string(issue.Body),
				URL:              string(issue.URL),
			}
			helpLookingRepo.Issues = append(helpLookingRepo.Issues, helpWantedIssue)
		}
		helpLookingRepos = append(helpLookingRepos, helpLookingRepo)
	}

	return helpLookingRepos
}
