package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"help-the-stars/internal/persistence"

	"github.com/charmbracelet/log"
)

// maxNumOfIssues is the maximum number of issues to notify at once.
const maxNumOfIssues = 7
const checkIntervalInMs = 3000

// DataController is a controller for data operations.
type DataController struct {
	queries         *persistence.Queries
	matrixClient    *MatrixClient
	settingsService *SettingsService
	ghStarsService  *GhStarsService
}

// CreateController creates a new DataController instance.
func CreateController(database *sql.DB,
	matrixClient *MatrixClient,
	settingsService *SettingsService) *DataController {
	ghStarsService := NewGithubStarService(settingsService)

	return &DataController{
		queries:         persistence.New(database),
		matrixClient:    matrixClient,
		settingsService: settingsService,
		ghStarsService:  ghStarsService,
	}
}

// GetDataForView retrieves data for the view.
func (d *DataController) GetDataForView(ctx context.Context) (ThankStarsData, error) {
	repos, err := d.queries.ListRepos(ctx)
	if err != nil {
		return ThankStarsData{}, fmt.Errorf("failed to list repos: %w", err)
	}

	issues, err := d.queries.ListIssues(ctx)
	if err != nil {
		return ThankStarsData{}, fmt.Errorf("failed to list issues: %w", err)
	}

	taskData, err := d.queries.GetTaskData(ctx)
	if err != nil {
		return ThankStarsData{}, fmt.Errorf("failed to get task data: %w", err)
	}

	return mapDbResultToViewModel(issues, repos, taskData), nil
}

// GetLastRun retrieves the last run time.
func (d *DataController) GetLastRun(ctx context.Context) (string, error) {
	taskData, err := d.queries.GetTaskData(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get task data: %w", err)
	}

	if taskData.LastRun.Valid {
		return taskData.LastRun.Time.String(), nil
	}
	// maybe first loading in progress
	return "", nil
}

// Worker starts the worker
// an endless loop checking for new data.
func (d *DataController) Worker() {
	log.Info("start worker...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initTaskData, err := d.queries.GetTaskData(ctx)
	if err != nil {
		log.Info("Init task data...")

		err = d.queries.InitTaskData(ctx)
	}

	if err != nil {
		log.Error("Error initializing task data", "error", err)

		return
	}

	if initTaskData.InProgress.Valid && initTaskData.InProgress.Bool {
		log.Warn("⚠️ Recuperating from bad stop")
	}

	for {
		taskData, err := d.queries.GetTaskData(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Error("Error retrieving task data", "error", err)

			return
		}

		d.checkAndDo(ctx, &taskData, err)

		log.Debug(".")
		time.Sleep(time.Duration(checkIntervalInMs) * time.Millisecond)
	}
}

func (d *DataController) checkAndDo(ctx context.Context, taskData *persistence.TaskDatum, err error) {
	if errors.Is(err, sql.ErrNoRows) ||
		!taskData.LastRun.Valid ||
		(taskData.LastRun.Valid &&
			time.Since(taskData.LastRun.Time) > time.Hour*time.Duration(d.settingsService.settings.Interval)) {
		log.Info("worker : time elapsed, get data...")

		if err := d.getAndSaveIssues(ctx); err != nil {
			d.matrixClient.NotifyError(ctx, err)
			// Updating task data to avoid short fail loop
			if err = d.queries.UpdateTimeTaskData(ctx,
				sql.NullTime{Time: time.Now(), Valid: true}); err != nil {
				log.Error("Error updating task data", "error", err)
			}

			return
		}
	}
}

func (d *DataController) getAndSaveIssues(ctx context.Context) error {
	if err := d.queries.TaskDataInProgress(ctx); err != nil {
		log.Error("Error getting starred repos", "error", err)

		return fmt.Errorf("failed to set task data in progress: %w", err)
	}

	log.Info("Loading issues...")

	repos, err := d.ghStarsService.GetStaredRepos(ctx)
	if err != nil {
		log.Error("Error listing issues from database", "error", err)

		return err
	}

	ghIssues := flattenIssues(repos)
	if ghIssues == nil {
		log.Warn("No issues found don't touch anything...")

		return nil
	}

	dbIssues, err := d.queries.ListIssues(ctx)
	if err != nil {
		log.Error("Error listing repos from database", "error", err)

		return fmt.Errorf("failed to list issues: %w", err)
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

	if err = d.queries.UpdateTimeTaskData(ctx,
		sql.NullTime{Time: time.Now(), Valid: true}); err != nil {
		log.Error("Error updating task data", "error", err)

		return fmt.Errorf("failed to update task data: %w", err)
	}

	return nil
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

		if delErr := d.queries.DeleteRepo(ctx, expiredRepos[i].RepoOwner); delErr != nil {
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

	if len(issues) > maxNumOfIssues {
		log.Info("Notify many issues")
		d.matrixClient.NotifySeveralNewIssues(ctx)

		return
	}

	for i := range issues {
		log.Info("Notify an issue " + issues[i].Url)
		d.matrixClient.Notify(ctx, &issues[i])
	}
}

// return new issue to notify and expired one to delete form base.
func sortNewAndExpired[T HasKey](incoming []T, issuesFromDb []T) (newOnes []T, expiredOnes []T) {
	for _, object := range incoming {
		found := false

		for _, issue := range issuesFromDb {
			if issue.Key() == object.Key() {
				found = true

				break
			}
		}

		if !found {
			newOnes = append(newOnes, object)
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
			expiredOnes = append(expiredOnes, issue)
		}
	}

	return newOnes, expiredOnes
}
