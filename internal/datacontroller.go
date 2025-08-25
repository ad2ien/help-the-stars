package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"help-the-stars/internal/persistence"
)

var internalSeconds = 3000

type DataController struct {
	queries *persistence.Queries
	ctx     context.Context
}

func CreateController(db *sql.DB) *DataController {
	ctx := context.Background()
	return &DataController{
		queries: persistence.New(db),
		ctx:     ctx,
	}
}

func (d *DataController) GetAndSaveIssues() {

	d.queries.TaskDataInProgress(d.ctx)

	// when, where from...
	fmt.Println("Loading issues...")
	data, err := GetStaredRepos(50)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(data); i++ {

		fmt.Println("Save an issue ", data[i].Url)
		d.queries.CreateIssue(d.ctx,
			mapModelToDbParameter(data[i]))
	}

	err = d.queries.UpdateTimeTaskData(d.ctx, sql.NullTime{Time: time.Now(), Valid: true})
	if err != nil {
		log.Fatal(err)
	}
}

func (d *DataController) Worker() {
	fmt.Print("start worker...")

	initTaskData, err := d.queries.GetTaskData(d.ctx)
	if err != nil {
		fmt.Println("Init task data...")
		err2 := d.queries.InitTaskData(d.ctx)
		if err2 != nil {
			log.Fatal(err2)
		}
	}
	if initTaskData.InProgress.Valid && initTaskData.InProgress.Bool {
		fmt.Println("⚠️ Recuperating from bad stop ")
	}

	for {
		taskData, err := d.queries.GetTaskData(d.ctx)
		if err != nil {
			log.Fatal(err)

		} else if !taskData.LastRun.Valid ||
			(taskData.LastRun.Valid && time.Since(taskData.LastRun.Time) > time.Hour*24) {
			fmt.Println("worker : time elapsed, get data...")
			d.GetAndSaveIssues()

		} else {
			fmt.Print(".")
			time.Sleep(time.Duration(internalSeconds) * time.Millisecond)
		}
	}
}
