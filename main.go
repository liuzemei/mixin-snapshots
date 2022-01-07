package main

import (
	"context"
	"time"
)

func main() {
	ctx := context.Background()
	initDB(ctx)
	go startGetSnapshotJobs(ctx)
	// go startDailyJob(ctx)
	// go startMonthJob(ctx)
	for {
		time.Sleep(time.Second)
		statisticDaily(ctx)
		statisticMonth(ctx)
	}
}
