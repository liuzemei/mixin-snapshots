package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

var _startAt, _ = time.Parse("2006-01-02", "2017-12-23")

type Asset struct {
	AssetID string          `json:"asset_id"`
	Amount  decimal.Decimal `json:"amount"`
	Date    time.Time       `json:"date"`
}

func statisticDaily(ctx context.Context) {
	// 1. 获取已经统计完了的上上一天的资产详情
	for {
		preAt := _startAt
		_ = pool.QueryRow(ctx, `SELECT date FROM asset_daily ORDER BY date DESC LIMIT 1`).Scan(&preAt)
		// 2. 获取最近这一条记录的最近一天的所有记录
		startAt := preAt.Add(time.Hour * 24)
		if startAt.Equal(GetZeroTimeByDay(time.Now())) {
			log.Printf("finished...%s\n", startAt)
			return
		}

		lastAt, err := getLastSnapshot(ctx)
		if err != nil {
			log.Println("statisticDaily 0", err)
			return
		}
		lastAt = GetZeroTimeByDay(lastAt)
		if startAt.Equal(lastAt) {
			return
		}

		rows, err := pool.Query(ctx, `SELECT asset_id, amount, date FROM asset_daily WHERE date=$1`, preAt)
		if err != nil {
			log.Println("statisticDaily 1", err)
			return
		}
		assetMap := make(map[string]decimal.Decimal)
		for rows.Next() {
			var asset Asset
			if err := rows.Scan(&asset.AssetID, &asset.Amount, &asset.Date); err != nil {
				log.Println("statisticDaily 2", err)
				return
			}
			assetMap[asset.AssetID] = asset.Amount
		}
		rows.Close()
		ss, err := getSnapshotsByDay(ctx, startAt)
		if err != nil {
			log.Println("statisticDaily 3", err)
			return
		}
		for _, v := range ss {
			if _, ok := assetMap[v.AssetID]; ok {
				assetMap[v.AssetID] = assetMap[v.AssetID].Add(v.Amount)
			} else {
				assetMap[v.AssetID] = v.Amount
			}
		}
		// 3. 保存到数据库
		if err := RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
			for k, v := range assetMap {
				_, err := tx.Exec(ctx, `
INSERT INTO asset_daily (asset_id, amount, date) 
VALUES ($1, $2, $3)`, k, v, startAt)
				if err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			log.Println("statisticDaily 4", err)
		}
	}
}

func statisticMonth(ctx context.Context) {
	for {
		// 1. 获取上一个月的最后一天的资产详情
		preAt := GetMonthStartByDay(_startAt)
		_ = pool.QueryRow(ctx, `SELECT date FROM asset_month ORDER BY date DESC LIMIT 1`).Scan(&preAt)
		startAt := preAt.AddDate(0, 1, 0)
		if startAt.Equal(GetMonthStartByDay(time.Now())) {
			log.Printf("finished...%s\n", startAt)
			return
		}
		var lastAt time.Time
		// lastAt, err := getLastSnapshot(ctx)
		err := pool.QueryRow(ctx, `SELECT date FROM asset_daily ORDER BY date DESC LIMIT 1`).Scan(&lastAt)
		if err != nil {
			log.Println("statisticDaily 0", err)
			return
		}
		lastAt = GetMonthStartByDay(lastAt)
		if startAt.Equal(lastAt) {
			return
		}
		targetAt := startAt.AddDate(0, 1, -1)
		rows, err := pool.Query(ctx, `SELECT asset_id, amount, date FROM asset_daily WHERE date=$1`, targetAt)
		if err != nil {
			log.Println("statisticMonth 1", err)
			return
		}
		assetMap := make(map[string]decimal.Decimal)
		for rows.Next() {
			var asset Asset
			if err := rows.Scan(&asset.AssetID, &asset.Amount, &asset.Date); err != nil {
				log.Println("statisticMonth 2", err)
				return
			}
			assetMap[asset.AssetID] = asset.Amount
		}
		rows.Close()
		RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
			for k, v := range assetMap {
				_, err := tx.Exec(ctx, `
INSERT INTO asset_month (asset_id, amount, date) 
VALUES ($1, $2, $3)`, k, v, startAt)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
}

func startDailyJob(ctx context.Context) {
	c := cron.New(cron.WithLocation(time.UTC))
	_, err := c.AddFunc("0 1 * * *", func() {
		log.Println("start daily job")
		statisticDaily(ctx)
	})
	if err != nil {
		SendMsgToDeveloper(ctx, "定时任务StartDailyJob。。。出问题了。。。"+err.Error())
		return
	}
	c.Start()
}

func startMonthJob(ctx context.Context) {
	c := cron.New(cron.WithLocation(time.UTC))
	_, err := c.AddFunc("0 2 1 * *", func() {
		log.Println("start month job")
		statisticMonth(ctx)
	})
	if err != nil {
		SendMsgToDeveloper(ctx, "定时任务StartMonthJob。。。出问题了。。。"+err.Error())
		return
	}
	c.Start()
}
