package internal

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"help-the-stars/internal/persistence"

	"github.com/charmbracelet/log"
)

const MAX_ISSUE_NOTIFS = 7
const CHECK_INTERVAL_S = 3000

type DataController struct {
	queries      *persistence.Queries
	ctx          context.Context
	matrixClient *MatrixClient
}

func CreateController(db *sql.DB, matrixClient *MatrixClient) *DataController {
	ctx := context.Background()
	return &DataController{
		queries:      persistence.New(db),
		ctx:          ctx,
		matrixClient: matrixClient,
	}
}

// TODO it would be nice to cache this
func (d *DataController) GetDataForView() (ThankStarsData, error) {
	repos, err := d.queries.ListRepos(d.ctx)
	if err != nil {
		return ThankStarsData{}, err
	}
	issues, err := d.queries.ListIssues(d.ctx)
	if err != nil {
		return ThankStarsData{}, err
	}
	taskData, err := d.queries.GetTaskData(d.ctx)
	if err != nil {
		return ThankStarsData{}, err
	}
	return mapDbResultToViewModel(issues, repos, taskData), nil
}

func (d *DataController) Worker() {
	log.Info("start worker...")

	initTaskData, err := d.queries.GetTaskData(d.ctx)
	if err != nil {
		log.Info("Init task data...")
		err = d.queries.InitTaskData(d.ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
	if initTaskData.InProgress.Valid && initTaskData.InProgress.Bool {
		log.Info("⚠️ Recuperating from bad stop")
	}

	for {
		taskData, err := d.queries.GetTaskData(d.ctx)
		if err != nil && err != sql.ErrNoRows {
			log.Fatal(err)
		}

		if err == sql.ErrNoRows ||
			!taskData.LastRun.Valid ||
			(taskData.LastRun.Valid &&
				time.Since(taskData.LastRun.Time) > time.Hour*time.Duration(GetSettings().Interval)) {
			log.Info("worker : time elapsed, get data...")
			d.GetAndSaveIssues()

		} else {
			log.Debug(".")
			time.Sleep(time.Duration(CHECK_INTERVAL_S) * time.Millisecond)
		}
	}
}

func (d *DataController) GetAndSaveIssues() {

	err := d.queries.TaskDataInProgress(d.ctx)

	if err != nil {
		log.Fatal(err)
	}

	log.Info("Loading issues...")
	repos, err := GetStaredRepos()
	if err != nil {
		log.Fatal(err)
	}

	ghIssues := flattenIssues(repos)
	if ghIssues == nil {
		log.Warn("No issues found don't touch anything...")
		return
	}

	dbIssues, err := d.queries.ListIssues(d.ctx)
	if err != nil {
		log.Fatal(err)
	}
	issuesFromDb := mapDbIssuesToViewIssues(dbIssues)

	dbRepos, err := d.queries.ListRepos(d.ctx)
	if err != nil {
		log.Fatal(err)
	}
	reposFromDb := mapDbReposToViewRepos(dbRepos)

	newIssues, expiredIssues := sortNewAndExpired(ghIssues, issuesFromDb)
	newRepos, expiredRepos := sortNewAndExpired(repos, reposFromDb)

	d.deleteIssues(expiredIssues)
	d.deleteRepos(expiredRepos)

	d.createRepos(newRepos)
	d.createIssues(newIssues)

	d.handleNotifications(newIssues)

	err = d.queries.UpdateTimeTaskData(d.ctx, sql.NullTime{Time: time.Now(), Valid: true})
	if err != nil {
		log.Fatal(err)
	}
}

func (d *DataController) deleteIssues(expiredIssues []HelpWantedIssue) {
	for i := range expiredIssues {
		log.Info("Delete an issue ", "url", expiredIssues[i].Url)
		delErr := d.queries.DeleteIssue(d.ctx, expiredIssues[i].Url)
		if delErr != nil {
			log.Error("Error deleting issue", "error", delErr)
		}
	}
}

func (d *DataController) createRepos(repos []Repo) {
	for _, r := range repos {

		log.Info("Save a repo " + r.RepoOwner)
		_, createErr := d.queries.CreateRepo(d.ctx,
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

func (d *DataController) createIssues(issues []HelpWantedIssue) {
	for _, i := range issues {
		log.Info("Save an issue " + i.Url)
		_, createErr := d.queries.CreateIssue(d.ctx,
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

func (d *DataController) deleteRepos(expiredRepos []Repo) {
	for i := range expiredRepos {
		log.Info("Delete repo", "repo", expiredRepos[i].RepoOwner)
		delErr := d.queries.DeleteRepo(d.ctx, expiredRepos[i].RepoOwner)
		if delErr != nil {
			log.Error("Error deleting issue", "error", delErr)
		}
	}
}

func (d *DataController) handleNotifications(issues []HelpWantedIssue) {
	if d.matrixClient == nil {
		return
	}
	if issues == nil {
		return
	}
	if len(issues) > MAX_ISSUE_NOTIFS {
		log.Info("Notify many issues")
		d.matrixClient.NotifySeveralNewIssues()
		return
	}
	for i := range issues {
		log.Info("Notify an issue " + issues[i].Url)
		d.matrixClient.Notify(&issues[i])
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
