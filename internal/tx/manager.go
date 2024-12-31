package tx

import (
	"context"
	"log/slog"

	logutil "github.com/RyanBard/go-log-util/pkg"
	"github.com/jmoiron/sqlx"
)

type txmgr struct {
	log *slog.Logger
	db  *sqlx.DB
}

func NewTXMGR(log *slog.Logger, db *sqlx.DB) *txmgr {
	return &txmgr{
		log: log.With(logutil.LogAttrSVC("TXManager")),
		db:  db,
	}
}

func (m txmgr) Do(ctx context.Context, joinTX *sqlx.Tx, f func(*sqlx.Tx) error) (err error) {
	log := m.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Do"),
	)
	var tx *sqlx.Tx
	if joinTX == nil {
		log.Debug("creating tx")
		tx, err = m.db.Beginx()
		if err != nil {
			log.With(logutil.LogAttrError(err)).Error("failed to create tx")
			return err
		}
	} else {
		log.Debug("joining tx")
		tx = joinTX
	}
	defer func() {
		if joinTX != nil {
			log.Debug("skipping commit/rollback of joined tx")
			return
		}
		if err != nil {
			log.With(logutil.LogAttrError(err)).Warn("f errored, rolling back tx")
			rbErr := tx.Rollback()
			if rbErr != nil {
				log.With(logutil.LogAttrError(rbErr)).Error("rollback failed")
			}
			return
		}
		log.Debug("f succeeded, committing tx")
		err = tx.Commit()
		if err != nil {
			log.With(logutil.LogAttrError(err)).Error("failed to commit tx")
		}
	}()
	log.Debug("calling f")
	err = f(tx)
	return err
}
