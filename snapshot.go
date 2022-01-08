package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Snapshot struct {
	SnapshotID string          `json:"snapshot_id"`
	Type       string          `json:"type"`
	Amount     decimal.Decimal `json:"amount"`
	AssetID    string          `json:"asset_id"`
	Source     string          `json:"source"`
	CreatedAt  time.Time       `json:"created_at"`
}

func startGetSnapshotJobs(ctx context.Context) {
	for {
		time.Sleep(time.Second * 5)
		getNewSnapshotAndSave(ctx)
	}
}

// 1. 获取
func getNewSnapshotAndSave(ctx context.Context) {
	// 1. 获取最近的一条记录
	lastSnapshotCreatedAt := _startAt
	if err := pool.QueryRow(ctx, `
SELECT created_at FROM snapshots 
ORDER BY created_at DESC LIMIT 1`).Scan(&lastSnapshotCreatedAt); err != nil {
		log.Println("getNewSnapshot 1", err)
	}
	// 2. 获取最近的这条记录之后的一些记录
	ss, err := mClient.ReadNetworkSnapshots(ctx, "", lastSnapshotCreatedAt, "ASC", 500)
	if err == nil {
		for _, s := range ss {
			if _, err := pool.Exec(ctx, `
INSERT INTO snapshots (snapshot_id, type, amount, asset_id, source, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`,
				s.SnapshotID, s.Type, s.Amount, s.AssetID, s.Source, s.CreatedAt); err != nil {
				if !strings.Contains(err.Error(), "duplicate key") {
					log.Println("getNewSnapshot 2", err)
					return
				}
			}
		}
	} else {
		log.Println("getNewSnapshot 3", err)
	}
}

func getSnapshotsByDay(ctx context.Context, startAt time.Time) ([]*Snapshot, error) {
	res := make([]*Snapshot, 0)
	endAt := startAt.Add(time.Hour * 24)
	rows, err := pool.Query(ctx, `
SELECT snapshot_id,type,amount,asset_id,source,created_at 
FROM snapshots 
WHERE source IN ('DEPOSIT_CONFIRMED','WITHDRAWAL_INITIALIZED')
AND created_at>=$1 
AND created_at<$2
`, startAt, endAt)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var s Snapshot
		if err := rows.Scan(&s.SnapshotID, &s.Type, &s.Amount, &s.AssetID, &s.Source, &s.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, &s)
	}
	rows.Close()
	return res, nil
}

func getLastSnapshot(ctx context.Context) (time.Time, error) {
	var t time.Time
	if err := pool.QueryRow(ctx, `
SELECT created_at FROM snapshots ORDER BY created_at DESC LIMIT 1`).Scan(&t); err != nil {
		return t, err
	}
	return t, nil
}
