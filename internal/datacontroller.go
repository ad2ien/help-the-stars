package internal

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"help-the-stars/internal/persistence"

	"github.com/charmbracelet/log"
)

var internalSeconds = 3000

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

func (d *DataController) GetDataForView() (ThankStarsData, error) {
	issues, err := d.queries.ListIssues(d.ctx)
	if err != nil {
		return ThankStarsData{}, err
	}
	taskData, err := d.queries.GetTaskData(d.ctx)
	if err != nil {
		return ThankStarsData{}, err
	}
	return mapDbResultToViewData(issues, taskData), nil
}

func (d *DataController) Worker() {
	log.Info("start worker...")

	initTaskData, err := d.queries.GetTaskData(d.ctx)
	if err != nil {
		log.Info("Init task data...")
		err2 := d.queries.InitTaskData(d.ctx)
		if err2 != nil {
			log.Fatal(err2)
		}
	}
	if initTaskData.InProgress.Valid && initTaskData.InProgress.Bool {
		log.Info("⚠️ Recuperating from bad stop ")
	}

	for {
		taskData, err := d.queries.GetTaskData(d.ctx)
		if err != nil {
			log.Fatal(err)

		} else if !taskData.LastRun.Valid ||
			(taskData.LastRun.Valid && time.Since(taskData.LastRun.Time) > time.Hour*24) {
			log.Info("worker : time elapsed, get data...")
			d.GetAndSaveIssues()

		} else {
			fmt.Print(".")
			time.Sleep(time.Duration(internalSeconds) * time.Millisecond)
		}
	}
}

func (d *DataController) GetAndSaveIssues() {

	err := d.queries.TaskDataInProgress(d.ctx)

	if err != nil {
		log.Fatal(err)
	}

	log.Info("Loading issues...")
	data, err := GetStaredRepos(50)
	if err != nil {
		log.Fatal(err)
	}

	news, expired := d.sortNewAndExpired(data)

	for i := 0; i < len(expired); i++ {
		log.Info("Delete an issue ", expired[i].Url)
		delErr := d.queries.DeleteIssue(d.ctx, expired[i].Url)
		if delErr != nil {
			log.Error("Error deleting issue", delErr)
		}
	}

	if d.matrixClient != nil {
		for i := 0; i < len(news); i++ {
			log.Info("Notify an issue " + news[i].Url)
			d.matrixClient.Notify(&news[i])
		}
	}

	for i := 0; i < len(data); i++ {

		log.Info("Save an issue ", data[i].Url)
		_, createErr := d.queries.CreateIssue(d.ctx,
			mapModelToDbParameter(data[i]))

		if createErr != nil {
			log.Error("Error creating issue", createErr)
		}
	}

	err = d.queries.UpdateTimeTaskData(d.ctx, sql.NullTime{Time: time.Now(), Valid: true})
	if err != nil {
		log.Fatal(err)
	}
}

// return new issue to notify and expired one to delete form base
func (d *DataController) sortNewAndExpired(ghIssues []HelpWantedIssue) ([]HelpWantedIssue, []HelpWantedIssue) {
	issues, err := d.queries.ListIssues(d.ctx)
	if err != nil {
		log.Fatal(err)
	}
	issuesFromDb := mapIssuesDbToModels(issues)

	expired := []HelpWantedIssue{}
	new := []HelpWantedIssue{}

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
