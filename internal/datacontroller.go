package internal

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"help-the-stars/internal/persistence"

	"github.com/charmbracelet/log"
)

const MAX_ISSUE_NOTIFS = 7
const CHECK_INTERVAL_S = 3000

type DataController struct {
	queries      *persistence.Queries
	matrixClient *MatrixClient
}

func CreateController(db *sql.DB, matrixClient *MatrixClient) *DataController {
	return &DataController{
		queries:      persistence.New(db),
		matrixClient: matrixClient,
	}
}

// TODO it would be nice to cache this
func (d *DataController) GetDataForView(ctx context.Context) (ThankStarsData, error) {
	repos, err := d.queries.ListRepos(ctx)
	if err != nil {
		return ThankStarsData{}, err
	}
	issues, err := d.queries.ListIssues(ctx)
	if err != nil {
		return ThankStarsData{}, err
	}
	taskData, err := d.queries.GetTaskData(ctx)
	if err != nil {
		return ThankStarsData{}, err
	}
	return mapDbResultToViewModel(issues, repos, taskData), nil
}

func (d *DataController) GetLastRun(ctx context.Context) (string, error) {
	taskData, err := d.queries.GetTaskData(ctx)
	if err != nil {
		return "", err
	}
	if taskData.LastRun.Valid {
		return taskData.LastRun.Time.String(), nil
	}
	// maybe first loading in progress
	return "", nil
}

func (d *DataController) Worker() {
	log.Info("start worker...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initTaskData, err := d.queries.GetTaskData(ctx)
	if err != nil {
		log.Info("Init task data...")
		err = d.queries.InitTaskData(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
	if initTaskData.InProgress.Valid && initTaskData.InProgress.Bool {
		log.Info("⚠️ Recuperating from bad stop")
	}

	for {
		taskData, err := d.queries.GetTaskData(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Fatal(err)
		}

		if errors.Is(err, sql.ErrNoRows) ||
			!taskData.LastRun.Valid ||
			(taskData.LastRun.Valid &&
				time.Since(taskData.LastRun.Time) > time.Hour*time.Duration(GetSettings().Interval)) {
			log.Info("worker : time elapsed, get data...")
			d.GetAndSaveIssues(ctx)

		} else {
			log.Debug(".")
			time.Sleep(time.Duration(CHECK_INTERVAL_S) * time.Millisecond)
		}
	}
}

func (d *DataController) GetAndSaveIssues(ctx context.Context) {

	err := d.queries.TaskDataInProgress(ctx)

	if err != nil {
		log.Fatal(err)
	}

	log.Info("Loading issues...")
	repos, err := GetStaredRepos(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ghIssues := flattenIssues(repos)
	if ghIssues == nil {
		log.Warn("No issues found don't touch anything...")
		return
	}

	dbIssues, err := d.queries.ListIssues(ctx)
	if err != nil {
		log.Fatal(err)
	}
	issuesFromDb := mapDbIssuesToViewIssues(dbIssues)

	dbRepos, err := d.queries.ListRepos(ctx)
	if err != nil {
		log.Fatal(err)
	}
	reposFromDb := mapDbReposToViewRepos(dbRepos)

	newIssues, expiredIssues := sortNewAndExpired(ghIssues, issuesFromDb)
	newRepos, expiredRepos := sortNewAndExpired(repos, reposFromDb)

	d.deleteIssues(ctx, expiredIssues)
	d.deleteRepos(ctx, expiredRepos)

	d.createRepos(ctx, newRepos)
	d.createIssues(ctx, newIssues)

	d.handleNotifications(ctx, newIssues)

	err = d.queries.UpdateTimeTaskData(ctx, sql.NullTime{Time: time.Now(), Valid: true})
	if err != nil {
		log.Fatal(err)
	}
}

func (d *DataController) deleteIssues(ctx context.Context,
	expiredIssues []HelpWantedIssue) {
	for i := range expiredIssues {
		log.Info("Delete an issue ", "url", expiredIssues[i].Url)
		delErr := d.queries.DeleteIssue(ctx, expiredIssues[i].Url)
		if delErr != nil {
			log.Error("Error deleting issue", "error", delErr)
		}
	}
}

func (d *DataController) createRepos(ctx context.Context, repos []Repo) {
	for _, r := range repos {

		log.Info("Save a repo " + r.RepoOwner)
		_, createErr := d.queries.CreateRepo(ctx,
			mapModelToRepoDbParameter(r))

		if createErr != nil {
			if strings.Contains(createErr.Error(), "UNIQUE constraint") {
				log.Info("Repo already exists")
			} else {
				log.Error("Error creating repo", "error", createErr)
			}
		}
	}
}

func (d *DataController) createIssues(ctx context.Context, issues []HelpWantedIssue) {
	for _, i := range issues {
		log.Info("Save an issue " + i.Url)
		_, createErr := d.queries.CreateIssue(ctx,
			mapModelToIssueDbParameter(i))

		if createErr != nil {
			if strings.Contains(createErr.Error(), "UNIQUE constraint") {
				log.Info("Issue already exists")
			} else {
				log.Error("Error creating issue", "error", createErr)
			}
		}
	}
}

func (d *DataController) deleteRepos(ctx context.Context, expiredRepos []Repo) {
	for i := range expiredRepos {
		log.Info("Delete repo", "repo", expiredRepos[i].RepoOwner)
		delErr := d.queries.DeleteRepo(ctx, expiredRepos[i].RepoOwner)
		if delErr != nil {
			log.Error("Error deleting issue", "error", delErr)
		}
	}
}

func (d *DataController) handleNotifications(ctx context.Context,
	issues []HelpWantedIssue) {
	if d.matrixClient == nil {
		return
	}
	if issues == nil {
		return
	}
	if len(issues) > MAX_ISSUE_NOTIFS {
		log.Info("Notify many issues")
		d.matrixClient.NotifySeveralNewIssues(ctx)
		return
	}
	for i := range issues {
		log.Info("Notify an issue " + issues[i].Url)
		d.matrixClient.Notify(ctx, &issues[i])
	}
}

// return new issue to notify and expired one to delete form base
func sortNewAndExpired[T HasKey](incoming []T, issuesFromDb []T) (new []T, expired []T) {

	for _, object := range incoming {
		found := false
		for _, issue := range issuesFromDb {
			if issue.Key() == object.Key() {
				found = true
				break
			}
		}
		if !found {
			new = append(new, object)
		}
	}

	for _, issue := range issuesFromDb {
		found := false
		for _, object := range incoming {
			if issue.Key() == object.Key() {
				found = true
				break
			}
		}
		if !found {
			expired = append(expired, issue)
		}
	}
	return new, expired
}
