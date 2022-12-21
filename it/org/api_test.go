//go:build integration
// +build integration

package org

import (
	"context"
	"errors"
	"fmt"
	"github.com/RyanBard/gin-ex/it/config"
	"github.com/RyanBard/gin-ex/pkg/org"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const (
	sysOrgID = "a517c24e-9b5f-4e5a-b840-e4f70a74725f"
)

type orgClient interface {
	GetByID(ctx context.Context, id string) (org.Org, error)
	GetAll(ctx context.Context) ([]org.Org, error)
	SearchByName(ctx context.Context, name string) ([]org.Org, error)
	Save(ctx context.Context, input org.Org) (org.Org, error)
	Delete(ctx context.Context, input org.DeleteOrg) error
}

type info struct {
	config        config.Config
	orgsToCleanup map[string]org.DeleteOrg
	orgClient     orgClient
	log           logrus.FieldLogger
	reqID         string
}

func setupSuite(tb testing.TB) (*info, func(tb testing.TB)) {
	reqID := uuid.NewString()
	logger := logrus.StandardLogger()
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.WithError(err).Fatal("invalid config")
	}
	logLvl, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.WithError(err).Fatal("invalid log level")
	}
	logger.SetLevel(logLvl)
	log := logger.WithField("reqID", reqID)
	s := info{
		config:        cfg,
		orgsToCleanup: make(map[string]org.DeleteOrg),
		orgClient: org.NewClient(
			org.Config{
				BaseURL: cfg.BaseURL,
			},
			http.Client{},
			func(isRetry bool) (string, error) {
				return "foo", nil
			},
		),
		log:   log,
		reqID: reqID,
	}
	return &s, func(tb testing.TB) {
		for i, o := range s.orgsToCleanup {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("org-teardown-%s-%s", reqID, i))
			err := s.orgClient.Delete(ctx, o)
			if err != nil {
				log.WithError(err).WithField("org", o).Warn("failed to cleanup org")
			}
		}
	}
}

func (oi *info) addOrgToCleanup(o org.Org) {
	oi.orgsToCleanup[o.ID] = org.DeleteOrg{ID: o.ID, Version: o.Version}
}

func IgnoreTestOrgAPI(t *testing.T) {
	s, teardown := setupSuite(t)
	defer teardown(t)

	t.Run("GetByID", func(t *testing.T) {
		t.Run("Found", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getByID-valid-%s", s.reqID))
			o, err := s.orgClient.GetByID(ctx, sysOrgID)
			assert.Nil(t, err)
			assert.NotNil(t, o.ID)
			assert.Equal(t, sysOrgID, o.ID)
			assert.Equal(t, "System Org", o.Name)
			assert.Equal(t, true, o.IsSystem)
			assert.Equal(t, int64(1), o.Version)
			assert.LessOrEqual(t, o.CreatedAt, o.UpdatedAt)
		})

		t.Run("NotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getByID-not-found-%s", s.reqID))
			_, err := s.orgClient.GetByID(ctx, "will-not-find")
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 404, httpErr.StatusCode)
		})
	})

	t.Run("GetAll", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getByID-valid-%s", s.reqID))
		orgs, err := s.orgClient.GetAll(ctx)
		assert.Nil(t, err)
		assert.GreaterOrEqual(t, 1, len(orgs))
		found := false
		for _, o := range orgs {
			if o.ID == sysOrgID {
				found = true
				assert.NotNil(t, o.ID)
				assert.Equal(t, sysOrgID, o.ID)
				assert.Equal(t, "System Org", o.Name)
				assert.Equal(t, true, o.IsSystem)
				assert.Equal(t, int64(1), o.Version)
				assert.LessOrEqual(t, o.CreatedAt, o.UpdatedAt)
			}
		}
		assert.True(t, found)
	})

	t.Run("SearchByName", func(t *testing.T) {
		t.Run("Found", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("searchByName-valid-%s", s.reqID))
			orgs, err := s.orgClient.SearchByName(ctx, "System")
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, 1, len(orgs))
			found := false
			for _, o := range orgs {
				if o.ID == sysOrgID {
					found = true
					assert.NotNil(t, o.ID)
					assert.Equal(t, sysOrgID, o.ID)
					assert.Equal(t, "System Org", o.Name)
					assert.Equal(t, true, o.IsSystem)
					assert.Equal(t, int64(1), o.Version)
					assert.LessOrEqual(t, o.CreatedAt, o.UpdatedAt)
				}
			}
			assert.True(t, found)
		})
		t.Run("NotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("searchByName-not-found-%s", s.reqID))
			orgs, err := s.orgClient.SearchByName(ctx, "will-not-find")
			assert.Nil(t, err)
			assert.Len(t, orgs, 0)
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-valid-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			assert.NotNil(t, o.ID)
			assert.Equal(t, int64(1), o.Version)
			assert.LessOrEqual(t, o.CreatedAt, o.UpdatedAt)
		})

		t.Run("MissingName", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-missing-name-%s", s.reqID))
			_, err := s.orgClient.Save(ctx, org.Org{
				Desc: "Integration Test",
			})
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("MissingDescription", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-missing-description-%s", s.reqID))
			_, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
			})
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("SysOrg", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-sys-org-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name:     "Test-" + uuid.NewString(),
				Desc:     "Integration Test",
				IsSystem: true,
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			assert.NotNil(t, o.ID)
			assert.False(t, o.IsSystem)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-valid-setup-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-valid-%s", s.reqID))
			o.Name = o.Name + "-updated"
			o2, err := s.orgClient.Save(ctx, o)
			s.addOrgToCleanup(o2)
			assert.Nil(t, err)
			assert.Equal(t, o.Version+1, o2.Version)
			assert.Equal(t, o.CreatedAt, o2.CreatedAt)
			assert.Less(t, o.UpdatedAt, o2.UpdatedAt)
			assert.Less(t, o2.CreatedAt, o2.UpdatedAt)
		})

		t.Run("OptimisticLock", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-opt-lock-setup-1-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-opt-lock-setup-2-%s", s.reqID))
			o.Name = o.Name + "-updated"
			o2, err := s.orgClient.Save(ctx, o)
			s.addOrgToCleanup(o2)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-opt-lock-%s", s.reqID))
			o.Name = o.Name + "-updated-again"
			_, err = s.orgClient.Save(ctx, o)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 409, httpErr.StatusCode)
		})

		t.Run("MissingName", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-missing-name-setup-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-missing-name-%s", s.reqID))
			o.Name = ""
			_, err = s.orgClient.Save(ctx, o)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("MissingDescription", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-missing-description-setup-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-missing-description-%s", s.reqID))
			o.Desc = ""
			_, err = s.orgClient.Save(ctx, o)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("NotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-not-found-%s", s.reqID))
			_, err := s.orgClient.Save(ctx, org.Org{
				ID:      "will-not-find",
				Name:    "Test-" + uuid.NewString(),
				Desc:    "Integration Test",
				Version: 1,
			})
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 404, httpErr.StatusCode)
		})

		t.Run("SysOrg", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-sys-org-%s", s.reqID))
			_, err := s.orgClient.Save(ctx, org.Org{
				ID:      sysOrgID,
				Name:    "Test-" + uuid.NewString(),
				Desc:    "Integration Test",
				Version: 1,
			})
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-valid-setup-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-valid-%s", s.reqID))
			err = s.orgClient.Delete(ctx, org.DeleteOrg{ID: o.ID, Version: o.Version})
			assert.Nil(t, err)
		})

		t.Run("OptimisticLock", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-opt-lock-setup-1-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-opt-lock-setup-2-%s", s.reqID))
			o.Name = o.Name + "-updated"
			o2, err := s.orgClient.Save(ctx, o)
			s.addOrgToCleanup(o2)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-opt-lock-%s", s.reqID))
			err = s.orgClient.Delete(ctx, org.DeleteOrg{ID: o.ID, Version: o.Version})
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 409, httpErr.StatusCode)
		})

		t.Run("MissingVersion", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-missing-version-setup-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-vmissing-version-%s", s.reqID))
			err = s.orgClient.Delete(ctx, org.DeleteOrg{ID: o.ID})
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("NotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-not-found-%s", s.reqID))
			err := s.orgClient.Delete(ctx, org.DeleteOrg{ID: "will-not-find", Version: 1})
			assert.Nil(t, err)
		})

		t.Run("SysOrg", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-sys-org-%s", s.reqID))
			err := s.orgClient.Delete(ctx, org.DeleteOrg{ID: sysOrgID, Version: 1})
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})
	})
}
