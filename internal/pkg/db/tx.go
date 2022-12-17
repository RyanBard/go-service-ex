package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type txmgr struct {
	log logrus.FieldLogger
	db  *sqlx.DB
}

func NewTXMGR(log logrus.FieldLogger, db *sqlx.DB) *txmgr {
	return &txmgr{
		log: log.WithField("SVC", "TXManager"),
		db:  db,
	}
}

func (m txmgr) Do(ctx context.Context, joinTX *sqlx.Tx, f func(*sqlx.Tx) error) (err error) {
	log := m.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Do",
	})
	var tx *sqlx.Tx
	if joinTX == nil {
		log.Debug("creating tx")
		tx, err = m.db.Beginx()
		if err != nil {
			log.WithError(err).Error("failed to create tx")
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
			log.WithError(err).Error("f errored, rolling back tx")
			rbErr := tx.Rollback()
			if rbErr != nil {
				log.WithError(rbErr).Error("rollback failed")
			}
			return
		}
		log.Debug("f succeeded, committing tx")
		err = tx.Commit()
		if err != nil {
			log.WithError(err).Error("failed to commit tx")
		}
	}()
	log.Debug("calling f")
	err = f(tx)
	return err
}
