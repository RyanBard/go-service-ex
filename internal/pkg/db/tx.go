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

func (m txmgr) Do(ctx context.Context, f func(*sqlx.Tx) error) (err error) {
	log := m.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Do",
	})
	log.Debug("creating tx")
	tx, err := m.db.Beginx()
	if err != nil {
		log.WithError(err).Error("failed to create tx")
		return err
	}
	log.Debug("calling f")
	err = f(tx)
	if err != nil {
		log.WithError(err).Error("f errored, rolling back tx")
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.WithError(rollbackErr).Error("rollback failed")
		}
		return err
	}
	log.Debug("f succeeded, committing tx")
	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("failed to commit tx")
		return err
	}
	return nil
}
