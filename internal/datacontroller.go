package internal

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"help-the-stars/internal/persistence"

	"github.com/charmbracelet/log"
)

const CHECK_INTERVAL_S = 3000
const DATA_REFRESH_INTERVAL_H = 2
const MAX_ISSUE_NOTIFS = 7

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
		if err != nil {
			log.Fatal(err)
		} else if !taskData.LastRun.Valid ||
			(taskData.LastRun.Valid && time.Since(taskData.LastRun.Time) > time.Hour*DATA_REFRESH_INTERVAL_H) {
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

	issues := flattenIssues(repos)
	if issues == nil {
		log.Fatal("No issues found, something went wrong")
		return
	}
	news, expired := d.sortNewAndExpired(issues)

	for i := range expired {
		log.Info("Delete an issue ", "url", expired[i].Url)
		delErr := d.queries.DeleteIssue(d.ctx, expired[i].Url)
		if delErr != nil {
			log.Error("Error deleting issue", "error", delErr)
		}
	}

	d.handleNotifications(news)

	for _, r := range repos {

		// Repo creation
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

		// Issue creation
		for _, i := range r.Issues {
			log.Info("Save an issue " + i.Url)
			_, createErr := d.queries.CreateIssue(d.ctx,
				mapModelToIssueDbParameter(i, r))

			if createErr != nil {
				if strings.Contains(createErr.Error(), "UNIQUE constraint") {
					log.Info("Issue already exists")
				} else {
					log.Error("Error creating issue", "error", createErr)
				}
			}
		}
	}

	err = d.queries.UpdateTimeTaskData(d.ctx, sql.NullTime{Time: time.Now(), Valid: true})
	if err != nil {
		log.Fatal(err)
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
func (d *DataController) sortNewAndExpired(ghIssues []HelpWantedIssue) (new []HelpWantedIssue, expired []HelpWantedIssue) {
	issues, err := d.queries.ListIssues(d.ctx)
	if err != nil {
		log.Fatal(err)
	}
	issuesFromDb := mapDbIssuesToViewIssues(issues)

	for _, ghIssue := range ghIssues {
		found := false
		for _, issue := range issuesFromDb {
			if issue.Url == ghIssue.Url {
				found = true
				break
			}
		}
		if !found {
			new = append(new, ghIssue)
		}
	}

	for _, issue := range issuesFromDb {
		found := false
		for _, ghIssue := range ghIssues {
			if issue.Url == ghIssue.Url {
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
