package user

import (
	"context"
	"time"

	"github.com/RyanBard/gin-ex/internal/pkg/org"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type OrgSVC interface {
	GetByID(ctx context.Context, id string) (org.Org, error)
}

type UserDAO interface {
	GetByID(ctx context.Context, id string) (User, error)
	GetAll(ctx context.Context) ([]User, error)
	GetAllByOrgID(ctx context.Context, orgID string) ([]User, error)
	Create(ctx context.Context, tx *sqlx.Tx, u User) error
	Update(ctx context.Context, tx *sqlx.Tx, u User) (User, error)
	Delete(ctx context.Context, tx *sqlx.Tx, u DeleteUser) error
}

type TXManager interface {
	Do(ctx context.Context, tx *sqlx.Tx, f func(*sqlx.Tx) error) error
}

type Timer interface {
	Now() time.Time
}

type IDGenerator interface {
	GenID() string
}

type service struct {
	log    logrus.FieldLogger
	orgSVC OrgSVC
	dao    UserDAO
	txMGR  TXManager
	timer  Timer
	idGen  IDGenerator
}

func NewService(log logrus.FieldLogger, orgSVC OrgSVC, dao UserDAO, txMGR TXManager, timer Timer, idGen IDGenerator) *service {
	return &service{
		log:    log.WithField("SVC", "UserSVC"),
		orgSVC: orgSVC,
		dao:    dao,
		txMGR:  txMGR,
		timer:  timer,
		idGen:  idGen,
	}
}

func (s service) GetByID(ctx context.Context, id string) (User, error) {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetByID",
		"id":    id,
	})
	log.Debug("called")
	return s.dao.GetByID(ctx, id)
}

func (s service) GetAll(ctx context.Context) ([]User, error) {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetAll",
	})
	log.Debug("called")
	return s.dao.GetAll(ctx)
}

func (s service) GetAllByOrgID(ctx context.Context, orgID string) ([]User, error) {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetAll",
		"orgID": orgID,
	})
	log.Debug("called")
	return s.dao.GetAllByOrgID(ctx, orgID)
}

func (s service) Save(ctx context.Context, u User) (out User, err error) {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Save",
		"user":  u,
	})
	log.Debug("called")
	if u.ID != "" {
		userInDB, err := s.GetByID(ctx, u.ID)
		if err != nil {
			log.WithError(err).Error("couldn't find user to update")
			return out, err
		}
		if userInDB.IsSystem {
			err = CannotModifySysUserErr{ID: u.ID}
			return out, err
		}
	}
	orgInDB, err := s.orgSVC.GetByID(ctx, u.OrgID)
	if err != nil {
		log.WithError(err).Error("couldn't find org to associate")
		return out, err
	}
	if orgInDB.IsSystem {
		err = CannotAssociateSysOrgErr{UserID: u.ID, OrgID: u.OrgID}
		return out, err
	}
	err = s.txMGR.Do(ctx, nil, func(tx *sqlx.Tx) error {
		if u.ID == "" {
			u.ID = s.idGen.GenID()
			u.Version = 1
			u.CreatedAt = s.timer.Now()
			u.CreatedBy = "TODO"
			u.UpdatedAt = s.timer.Now()
			u.UpdatedBy = "TODO"
			out = u
			return s.dao.Create(ctx, tx, u)
		} else {
			u.UpdatedAt = s.timer.Now()
			u.UpdatedBy = "TODO"
			out, err = s.dao.Update(ctx, tx, u)
			return err
		}
	})
	if err != nil {
		return User{}, err
	}
	return out, nil
}

func (s service) Delete(ctx context.Context, u DeleteUser) error {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Delete",
		"u":     u,
	})
	log.Debug("called")
	userInDB, err := s.GetByID(ctx, u.ID)
	if err != nil {
		log.WithError(err).Error("couldn't find user to delete")
		return err
	}
	if userInDB.IsSystem {
		err = CannotModifySysUserErr{ID: u.ID}
		return err
	}
	err = s.txMGR.Do(ctx, nil, func(tx *sqlx.Tx) error {
		return s.dao.Delete(ctx, tx, u)
	})
	if err != nil {
		return err
	}
	return nil
}
