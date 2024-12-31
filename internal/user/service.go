package user

import (
	"context"
	"errors"
	"log/slog"
	"time"

	ctxutil "github.com/RyanBard/go-ctx-util/pkg"
	logutil "github.com/RyanBard/go-log-util/pkg"
	"github.com/RyanBard/go-service-ex/pkg/org"
	"github.com/RyanBard/go-service-ex/pkg/user"
	"github.com/jmoiron/sqlx"
)

type OrgSVC interface {
	GetByID(ctx context.Context, id string) (org.Org, error)
}

type UserDAO interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetAll(ctx context.Context) ([]user.User, error)
	GetAllByOrgID(ctx context.Context, orgID string) ([]user.User, error)
	Create(ctx context.Context, tx *sqlx.Tx, u user.User) error
	Update(ctx context.Context, tx *sqlx.Tx, u user.User) (user.User, error)
	Delete(ctx context.Context, tx *sqlx.Tx, u user.DeleteUser) error
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
	log    *slog.Logger
	orgSVC OrgSVC
	dao    UserDAO
	txMGR  TXManager
	timer  Timer
	idGen  IDGenerator
}

func NewService(log *slog.Logger, orgSVC OrgSVC, dao UserDAO, txMGR TXManager, timer Timer, idGen IDGenerator) *service {
	return &service{
		log:    log.With(logutil.LogAttrSVC("UserSVC")),
		orgSVC: orgSVC,
		dao:    dao,
		txMGR:  txMGR,
		timer:  timer,
		idGen:  idGen,
	}
}

func (s service) GetByID(ctx context.Context, id string) (user.User, error) {
	log := s.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetByID"),
		logAttrUserID(id),
	)
	log.Debug("called")
	return s.dao.GetByID(ctx, id)
}

func (s service) GetAll(ctx context.Context) ([]user.User, error) {
	log := s.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetAll"),
	)
	log.Debug("called")
	return s.dao.GetAll(ctx)
}

func (s service) GetAllByOrgID(ctx context.Context, orgID string) ([]user.User, error) {
	log := s.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetAllByOrgID"),
		logAttrOrgID(orgID),
	)
	log.Debug("called")
	return s.dao.GetAllByOrgID(ctx, orgID)
}

func (s service) Save(ctx context.Context, u user.User) (out user.User, err error) {
	loggedInUserID, _ := ctx.Value(ctxutil.ContextKeyUserID{}).(string)
	if loggedInUserID == "" {
		return out, errors.New("user not logged in")
	}
	log := s.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Save"),
		logAttrUser(u),
	)
	log.Debug("called")
	if u.ID != "" {
		userInDB, err := s.GetByID(ctx, u.ID)
		if err != nil {
			return out, err
		}
		if userInDB.IsSystem {
			err = CannotModifySysUserErr{ID: u.ID}
			return out, err
		}
	}
	orgInDB, err := s.orgSVC.GetByID(ctx, u.OrgID)
	if err != nil {
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
			u.CreatedBy = loggedInUserID
			u.UpdatedAt = s.timer.Now()
			u.UpdatedBy = loggedInUserID
			u.IsSystem = false
			out = u
			return s.dao.Create(ctx, tx, u)
		} else {
			u.UpdatedAt = s.timer.Now()
			u.UpdatedBy = loggedInUserID
			out, err = s.dao.Update(ctx, tx, u)
			return err
		}
	})
	if err != nil {
		return user.User{}, err
	}
	return out, nil
}

func (s service) Delete(ctx context.Context, u user.DeleteUser) error {
	log := s.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Delete"),
		logAttrUser(u),
	)
	log.Debug("called")
	userInDB, err := s.GetByID(ctx, u.ID)
	if err != nil {
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
