//go:build integration
// +build integration

package org

import (
	"context"
	"errors"
	"fmt"
	"github.com/RyanBard/gin-ex/it/config"
	"github.com/RyanBard/gin-ex/pkg/org"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

const (
	sysOrgID       = "a517c24e-9b5f-4e5a-b840-e4f70a74725f"
	sysUserID      = "fc83cf36-bba0-41f0-8125-2ebc03087140"
	adminUserID    = "fc83cf36-bba0-41f0-8125-2ebc03087140"
	nonAdminUserID = "ffff0000-0000-0000-0000-000000000000"
)

type orgClient interface {
	GetByID(ctx context.Context, id string) (org.Org, error)
	GetAll(ctx context.Context) ([]org.Org, error)
	SearchByName(ctx context.Context, name string) ([]org.Org, error)
	Save(ctx context.Context, input org.Org) (org.Org, error)
	Delete(ctx context.Context, input org.DeleteOrg) error
}

type info struct {
	config            config.Config
	orgsToCleanup     map[string]org.DeleteOrg
	orgClient         orgClient
	invJWTOrgClient   orgClient
	nonAdminOrgClient orgClient
	log               logrus.FieldLogger
	reqID             string
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
	oi := info{
		config:        cfg,
		orgsToCleanup: make(map[string]org.DeleteOrg),
		log:           log,
		reqID:         reqID,
	}
	oi.orgClient = org.NewClient(
		org.Config{
			BaseURL: cfg.BaseURL,
		},
		http.Client{},
		func(isRetry bool) (string, error) {
			token := oi.getAdminJWT()
			return token, nil
		},
	)
	oi.invJWTOrgClient = org.NewClient(
		org.Config{
			BaseURL: cfg.BaseURL,
		},
		http.Client{},
		func(isRetry bool) (string, error) {
			return "x.y.z", nil
		},
	)
	oi.nonAdminOrgClient = org.NewClient(
		org.Config{
			BaseURL: cfg.BaseURL,
		},
		http.Client{},
		func(isRetry bool) (string, error) {
			token := oi.getNonAdminJWT()
			return token, nil
		},
	)
	return &oi, func(tb testing.TB) {
		for i, o := range oi.orgsToCleanup {
			if o.ID != "" {
				ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("org-teardown-%s-%s", reqID, i))
				err := oi.orgClient.Delete(ctx, o)
				if err != nil {
					log.WithError(err).WithField("org", o).Warn("failed to cleanup org")
				}
			}
		}
	}
}

func (oi *info) addOrgToCleanup(o org.Org) {
	oi.orgsToCleanup[o.ID] = org.DeleteOrg{ID: o.ID, Version: o.Version}
}

func (oi *info) getAdminJWT() string {
	claims := oi.getClaims(adminUserID)
	claims["admin"] = true
	return oi.hmacJWT(claims)
}

func (oi *info) getNonAdminJWT() string {
	claims := oi.getClaims(nonAdminUserID)
	return oi.hmacJWT(claims)
}

func (oi *info) getClaims(userID string) jwt.MapClaims {
	return jwt.MapClaims{
		"sub": userID,
		"aud": oi.config.JWTAudience,
		"iss": oi.config.JWTIssuer,
		"exp": time.Now().AddDate(0, 0, 1).Unix(),
	}
}

func (oi *info) hmacJWT(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(oi.config.JWTSecret))
	if err != nil {
		oi.log.WithError(err).Fatal("failed to sign jwt")
	}
	return tokenStr
}

func TestOrgAPI(t *testing.T) {
	s, teardown := setupSuite(t)
	defer teardown(t)

	s.log.WithField("reqID", s.reqID).Info("User Integration Test run")

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

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getByID-non-admin-jwt-%s", s.reqID))
			o, err := s.nonAdminOrgClient.GetByID(ctx, sysOrgID)
			assert.Nil(t, err)
			assert.NotNil(t, o.ID)
			assert.Equal(t, sysOrgID, o.ID)
			assert.Equal(t, "System Org", o.Name)
			assert.Equal(t, true, o.IsSystem)
			assert.Equal(t, int64(1), o.Version)
			assert.LessOrEqual(t, o.CreatedAt, o.UpdatedAt)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getByID-invalid-jwt-%s", s.reqID))
			_, err := s.invJWTOrgClient.GetByID(ctx, sysOrgID)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
		})
	})

	t.Run("GetAll", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAll-valid-%s", s.reqID))
			orgs, err := s.orgClient.GetAll(ctx)
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(orgs), 1)
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

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAll-non-admin-jwt-%s", s.reqID))
			orgs, err := s.nonAdminOrgClient.GetAll(ctx)
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(orgs), 1)
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

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAll-invalid-jwt-%s", s.reqID))
			_, err := s.invJWTOrgClient.GetAll(ctx)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
		})
	})

	t.Run("SearchByName", func(t *testing.T) {
		t.Run("Found", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("searchByName-valid-%s", s.reqID))
			orgs, err := s.orgClient.SearchByName(ctx, "System")
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(orgs), 1)
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

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("searchByName-non-admin-jwt-%s", s.reqID))
			orgs, err := s.nonAdminOrgClient.SearchByName(ctx, "System")
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(orgs), 1)
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

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("searchByName-invalid-jwt-%s", s.reqID))
			_, err := s.invJWTOrgClient.SearchByName(ctx, "System")
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
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
			o, err := s.orgClient.Save(ctx, org.Org{
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("DuplicateName", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-dup-name-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "System Org",
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 409, httpErr.StatusCode)
		})

		t.Run("MissingDescription", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-missing-description-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
			})
			s.addOrgToCleanup(o)
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

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-non-admin-jwt-%s", s.reqID))
			o, err := s.nonAdminOrgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-invalid-jwt-%s", s.reqID))
			o, err := s.invJWTOrgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
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
			o3, err := s.orgClient.Save(ctx, o)
			s.addOrgToCleanup(o3)
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
			o2, err := s.orgClient.Save(ctx, o)
			s.addOrgToCleanup(o2)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("DuplicateName", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-duplicate-name-setup-%s", s.reqID))
			name := "Test-" + uuid.NewString()
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: name,
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-dup-name-%s", s.reqID))
			o.Name = "System Org"
			o2, err := s.orgClient.Save(ctx, o)
			s.addOrgToCleanup(o2)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 409, httpErr.StatusCode)
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
			o2, err := s.orgClient.Save(ctx, o)
			s.addOrgToCleanup(o2)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("NotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-not-found-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				ID:      "will-not-find",
				Name:    "Test-" + uuid.NewString(),
				Desc:    "Integration Test",
				Version: 1,
			})
			s.addOrgToCleanup(o)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 404, httpErr.StatusCode)
		})

		t.Run("SysOrg", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-sys-org-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				ID:      sysOrgID,
				Name:    "Test-" + uuid.NewString(),
				Desc:    "Integration Test",
				Version: 1,
			})
			s.addOrgToCleanup(o)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-non-admin-jwt-setup-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-non-admin-jwt-%s", s.reqID))
			o.Desc = o.Desc + "-updated"
			o2, err := s.nonAdminOrgClient.Save(ctx, o)
			s.addOrgToCleanup(o2)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-invalid-jwt-setup-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-invalid-jwt-%s", s.reqID))
			o.Desc = o.Desc + "-updated"
			o2, err := s.invJWTOrgClient.Save(ctx, o)
			s.addOrgToCleanup(o2)
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
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

		t.Run("Idempotent", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-idempotent-setup-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-idempotent-1-%s", s.reqID))
			err = s.orgClient.Delete(ctx, org.DeleteOrg{ID: o.ID, Version: o.Version})
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-idempotent-2-%s", s.reqID))
			err = s.orgClient.Delete(ctx, org.DeleteOrg{ID: o.ID, Version: o.Version})
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-idempotent-3-%s", s.reqID))
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
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-missing-version-%s", s.reqID))
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

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-non-admin-jwt-setup-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-non-admin-jwt-%s", s.reqID))
			err = s.nonAdminOrgClient.Delete(ctx, org.DeleteOrg{ID: o.ID, Version: o.Version})
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-invalid-jwt-setup-%s", s.reqID))
			o, err := s.orgClient.Save(ctx, org.Org{
				Name: "Test-" + uuid.NewString(),
				Desc: "Integration Test",
			})
			s.addOrgToCleanup(o)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-invalid-jwt-%s", s.reqID))
			err = s.invJWTOrgClient.Delete(ctx, org.DeleteOrg{ID: o.ID, Version: o.Version})
			assert.NotNil(t, err)
			var httpErr org.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
		})
	})
}
