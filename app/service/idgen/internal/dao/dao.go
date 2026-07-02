/*
 * Created from 'scheme.tl' by 'mtprotoc'
 *
 * Copyright (c) 2021-present,  Teamgram Studio (https://teamgram.io).
 *  All rights reserved.
 *
 * Author: teamgramio (teamgram.io@gmail.com)
 */

package dao

import (
	"strconv"
	"context"
	"log"

	"github.com/bwmarrin/snowflake"
	"github.com/teamgram/marmota/pkg/stores/sqlx"
	"github.com/teamgram/teamgram-server/app/service/idgen/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/kv"
)

type Dao struct {
	*snowflake.Node
	KV  kv.Store
	db  *sqlx.DB
}

func New(c config.Config) *Dao {
	var (
		err error
		d   = new(Dao)
	)

	d.Node, err = snowflake.NewNode(c.NodeId)
	if err != nil {
		log.Fatal("new snowflake node error: ", err)
	}
	d.KV = kv.NewStore(c.SeqIDGen)

	// Initialize MySQL connection for counter sync
	if c.Mysql.DSN != "" {
		d.db = sqlx.NewMySQL(&c.Mysql)
		logx.Info("idgen: MySQL connection established for counter sync")
	}

	return d
}

// SyncCountersFromMySQL syncs Redis sequence IDs from MySQL max values on startup
func (d *Dao) SyncCountersFromMySQL(ctx context.Context) error {
	if d.db == nil {
		logx.Info("idgen: MySQL not configured, skipping counter sync")
		return nil
	}

	// Sync message_box_ngen_* counters from messages table
	if err := d.syncMessageBoxCounters(ctx); err != nil {
		logx.Errorf("idgen: sync message_box_ngen failed: %v", err)
		return err
	}

	// Sync pts_updates_ngen_* counters from user_pts_updates table
	if err := d.syncPtsCounters(ctx); err != nil {
		logx.Errorf("idgen: sync pts_updates_ngen failed: %v", err)
		return err
	}

	// Sync channel_message_box_ngen_* counters for channels (peer_type=3)
	if err := d.syncChannelMessageBoxCounters(ctx); err != nil {
		logx.Errorf("idgen: sync channel_message_box_ngen failed: %v", err)
		return err
	}

	logx.Info("idgen: all counters synced from MySQL successfully")
	return nil
}

func (d *Dao) syncMessageBoxCounters(ctx context.Context) error {
	// Get all user_ids that have messages
	rows, err := d.db.Query(ctx, `SELECT DISTINCT user_id FROM messages`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return err
		}

		// Get max user_message_box_id for this user
		var maxBoxID sqlx.NullInt64
		err = d.db.QueryRow(ctx, `SELECT MAX(user_message_box_id) FROM messages WHERE user_id = ?`, userID).
			Scan(&maxBoxID)
		if err != nil {
			return err
		}

		if maxBoxID.Valid && maxBoxID.Int64 > 0 {
			nextID := maxBoxID.Int64 + 1
			key := "message_box_ngen_" + string(rune(userID)) // Actually need strconv
			// Use proper string conversion
			key = "message_box_ngen_" + strconv.FormatInt(userID, 10)
			if err := d.KV.Set(ctx, key, strconv.FormatInt(nextID, 10)); err != nil {
				logx.Errorf("idgen: failed to set %s: %v", key, err)
				continue
			}
			logx.Infof("idgen: synced %s = %d (was %d)", key, nextID, maxBoxID.Int64)
		}
	}
	return rows.Err()
}

func (d *Dao) syncPtsCounters(ctx context.Context) error {
	rows, err := d.db.Query(ctx, `SELECT DISTINCT user_id FROM user_pts_updates`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return err
		}

		var maxPts sqlx.NullInt64
		err = d.db.QueryRow(ctx, `SELECT MAX(pts) FROM user_pts_updates WHERE user_id = ?`, userID).
			Scan(&maxPts)
		if err != nil {
			return err
		}

		if maxPts.Valid && maxPts.Int64 > 0 {
			nextID := maxPts.Int64 + 1
			key := "pts_updates_ngen_" + strconv.FormatInt(userID, 10)
			if err := d.KV.Set(ctx, key, strconv.FormatInt(nextID, 10)); err != nil {
				logx.Errorf("idgen: failed to set %s: %v", key, err)
				continue
			}
			logx.Infof("idgen: synced %s = %d (was %d)", key, nextID, maxPts.Int64)
		}
	}
	return rows.Err()
}

func (d *Dao) syncChannelMessageBoxCounters(ctx context.Context) error {
	rows, err := d.db.Query(ctx, `SELECT DISTINCT peer_id FROM messages WHERE peer_type = 3`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var channelID int64
		if err := rows.Scan(&channelID); err != nil {
			return err
		}

		var maxBoxID sqlx.NullInt64
		err = d.db.QueryRow(ctx, `SELECT MAX(user_message_box_id) FROM messages WHERE peer_type = 3 AND peer_id = ?`, channelID).
			Scan(&maxBoxID)
		if err != nil {
			return err
		}

		if maxBoxID.Valid && maxBoxID.Int64 > 0 {
			nextID := maxBoxID.Int64 + 1
			key := "channel_message_box_ngen_" + strconv.FormatInt(channelID, 10)
			if err := d.KV.Set(ctx, key, strconv.FormatInt(nextID, 10)); err != nil {
				logx.Errorf("idgen: failed to set %s: %v", key, err)
				continue
			}
			logx.Infof("idgen: synced %s = %d (was %d)", key, nextID, maxBoxID.Int64)
		}
	}
	return rows.Err()
}
