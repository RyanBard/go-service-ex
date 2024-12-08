//go:build integration
// +build integration

package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/RyanBard/go-service-ex/internal/apiclient"
	"github.com/RyanBard/go-service-ex/internal/httpx"
	"github.com/RyanBard/go-service-ex/it/config"
	"github.com/RyanBard/go-service-ex/pkg/org"
	"github.com/RyanBard/go-service-ex/pkg/user"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	sysOrgID       = "a517c24e-9b5f-4e5a-b840-e4f70a74725f"
	sysUserID      = "fc83cf36-bba0-41f0-8125-2ebc03087140"
	adminUserID    = "fc83cf36-bba0-41f0-8125-2ebc03087140"
	nonAdminUserID = "ffff0000-0000-0000-0000-000000000000"
)

type orgClient interface {
	Save(ctx context.Context, input org.Org) (org.Org, error)
	Delete(ctx context.Context, input org.DeleteOrg) error
}

type userClient interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetAll(ctx context.Context) ([]user.User, error)
	GetAllByOrgID(ctx context.Context, orgID string) ([]user.User, error)
	Save(ctx context.Context, input user.User) (user.User, error)
	Delete(ctx context.Context, input user.DeleteUser) error
}

type info struct {
	config             config.Config
	usersToCleanup     map[string]user.DeleteUser
	testOrg            org.Org
	orgClient          orgClient
	userClient         userClient
	invJWTUserClient   userClient
	nonAdminUserClient userClient
	log                logrus.FieldLogger
	reqID              string
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

	ui := info{
		config:         cfg,
		usersToCleanup: make(map[string]user.DeleteUser),
		log:            log,
		reqID:          reqID,
	}
	ui.orgClient = org.NewClient(
		org.Config{
			BaseURL: cfg.BaseURL,
		},
		apiclient.NewClient(
			httpx.NewClient(http.Client{}),
			func(isRetry bool) (string, error) {
				return ui.getAdminJWT(), nil
			},
		),
	)
	ui.userClient = user.NewClient(
		user.Config{
			BaseURL: cfg.BaseURL,
		},
		apiclient.NewClient(
			httpx.NewClient(http.Client{}),
			func(isRetry bool) (string, error) {
				return ui.getAdminJWT(), nil
			},
		),
	)
	ui.invJWTUserClient = user.NewClient(
		user.Config{
			BaseURL: cfg.BaseURL,
		},
		apiclient.NewClient(
			httpx.NewClient(http.Client{}),
			func(isRetry bool) (string, error) {
				return "x.y.z", nil
			},
		),
	)
	ui.nonAdminUserClient = user.NewClient(
		user.Config{
			BaseURL: cfg.BaseURL,
		},
		apiclient.NewClient(
			httpx.NewClient(http.Client{}),
			func(isRetry bool) (string, error) {
				return ui.getNonAdminJWT(), nil
			},
		),
	)

	ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("user-suite-setup-%s", reqID))
	testOrg, err := ui.orgClient.Save(ctx, org.Org{
		Name: "Test-" + uuid.NewString(),
		Desc: "Integration Test",
	})
	if err != nil {
		log.WithError(err).Fatal("failed to create test org")
	}
	ui.testOrg = testOrg

	return &ui, func(tb testing.TB) {
		for i, u := range ui.usersToCleanup {
			if u.ID != "" {
				ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("user-teardown-%s-%s", reqID, i))
				err := ui.userClient.Delete(ctx, u)
				if err != nil {
					log.WithError(err).WithField("user", u).Warn("failed to cleanup user")
				}
			}
		}
		ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("user-teardown-%s", reqID))
		o := org.DeleteOrg{ID: ui.testOrg.ID, Version: ui.testOrg.Version}
		err := ui.orgClient.Delete(ctx, o)
		if err != nil {
			log.WithError(err).WithField("org", o).Warn("failed to cleanup org")
		}
	}
}

func (ui *info) addUserToCleanup(u user.User) {
	ui.usersToCleanup[u.ID] = user.DeleteUser{ID: u.ID, Version: u.Version}
}

func (ui *info) getAdminJWT() string {
	claims := ui.getClaims(adminUserID)
	claims["admin"] = true
	return ui.hmacJWT(claims)
}

func (ui *info) getNonAdminJWT() string {
	claims := ui.getClaims(nonAdminUserID)
	return ui.hmacJWT(claims)
}

func (ui *info) getClaims(userID string) jwt.MapClaims {
	return jwt.MapClaims{
		"sub": userID,
		"aud": ui.config.JWTAudience,
		"iss": ui.config.JWTIssuer,
		"exp": time.Now().AddDate(0, 0, 1).Unix(),
	}
}

func (ui *info) hmacJWT(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(ui.config.JWTSecret))
	if err != nil {
		ui.log.WithError(err).Fatal("failed to sign jwt")
	}
	return tokenStr
}

func TestUserAPI(t *testing.T) {
	s, teardown := setupSuite(t)
	defer teardown(t)

	s.log.WithField("reqID", s.reqID).Info("User Integration Test run")

	t.Run("GetByID", func(t *testing.T) {
		t.Run("Found", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getByID-valid-%s", s.reqID))
			u, err := s.userClient.GetByID(ctx, sysUserID)
			assert.Nil(t, err)
			assert.NotNil(t, u.ID)
			assert.Equal(t, sysUserID, u.ID)
			assert.Equal(t, "System Admin", u.Name)
			assert.Equal(t, true, u.IsSystem)
			assert.Equal(t, true, u.IsAdmin)
			assert.Equal(t, true, u.IsActive)
			assert.Equal(t, int64(1), u.Version)
			assert.LessOrEqual(t, u.CreatedAt, u.UpdatedAt)
		})

		t.Run("NotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getByID-not-found-%s", s.reqID))
			_, err := s.userClient.GetByID(ctx, "will-not-find")
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 404, httpErr.StatusCode)
		})

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getByID-non-admin-jwt-%s", s.reqID))
			u, err := s.nonAdminUserClient.GetByID(ctx, sysUserID)
			assert.Nil(t, err)
			assert.NotNil(t, u.ID)
			assert.Equal(t, sysUserID, u.ID)
			assert.Equal(t, "System Admin", u.Name)
			assert.Equal(t, true, u.IsSystem)
			assert.Equal(t, true, u.IsAdmin)
			assert.Equal(t, true, u.IsActive)
			assert.Equal(t, int64(1), u.Version)
			assert.LessOrEqual(t, u.CreatedAt, u.UpdatedAt)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getByID-invalid-jwt-%s", s.reqID))
			_, err := s.invJWTUserClient.GetByID(ctx, sysUserID)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
		})
	})

	t.Run("GetAll", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAll-valid-%s", s.reqID))
			users, err := s.userClient.GetAll(ctx)
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(users), 1)
			found := false
			for _, u := range users {
				if u.ID == sysUserID {
					found = true
					assert.NotNil(t, u.ID)
					assert.Equal(t, sysUserID, u.ID)
					assert.Equal(t, sysOrgID, u.OrgID)
					assert.Equal(t, "System Admin", u.Name)
					assert.Equal(t, true, u.IsSystem)
					assert.Equal(t, true, u.IsAdmin)
					assert.Equal(t, true, u.IsActive)
					assert.Equal(t, int64(1), u.Version)
					assert.LessOrEqual(t, u.CreatedAt, u.UpdatedAt)
				}
			}
			assert.True(t, found)
		})

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAll-non-admin-jwt-%s", s.reqID))
			users, err := s.nonAdminUserClient.GetAll(ctx)
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(users), 1)
			found := false
			for _, u := range users {
				if u.ID == sysUserID {
					found = true
					assert.NotNil(t, u.ID)
					assert.Equal(t, sysUserID, u.ID)
					assert.Equal(t, sysOrgID, u.OrgID)
					assert.Equal(t, "System Admin", u.Name)
					assert.Equal(t, true, u.IsSystem)
					assert.Equal(t, true, u.IsAdmin)
					assert.Equal(t, true, u.IsActive)
					assert.Equal(t, int64(1), u.Version)
					assert.LessOrEqual(t, u.CreatedAt, u.UpdatedAt)
				}
			}
			assert.True(t, found)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAll-invalid-jwt-%s", s.reqID))
			_, err := s.invJWTUserClient.GetAll(ctx)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
		})
	})

	t.Run("GetAllByOrgID", func(t *testing.T) {
		t.Run("Found", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAllByOrgID-valid-%s", s.reqID))
			users, err := s.userClient.GetAllByOrgID(ctx, sysOrgID)
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(users), 1)
			found := false
			for _, u := range users {
				if u.ID == sysUserID {
					found = true
					assert.NotNil(t, u.ID)
					assert.Equal(t, sysUserID, u.ID)
					assert.Equal(t, sysOrgID, u.OrgID)
					assert.Equal(t, "System Admin", u.Name)
					assert.Equal(t, true, u.IsSystem)
					assert.Equal(t, true, u.IsAdmin)
					assert.Equal(t, true, u.IsActive)
					assert.Equal(t, int64(1), u.Version)
					assert.LessOrEqual(t, u.CreatedAt, u.UpdatedAt)
				}
			}
			assert.True(t, found)
		})

		t.Run("Empty", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAllByOrgID-empty-%s", s.reqID))
			users, err := s.userClient.GetAllByOrgID(ctx, s.testOrg.ID)
			assert.Nil(t, err)
			assert.Len(t, users, 0)
		})

		t.Run("OrgNotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAllByOrgID-org-not-found-%s", s.reqID))
			users, err := s.userClient.GetAllByOrgID(ctx, "will-not-find")
			assert.Nil(t, err)
			assert.Len(t, users, 0)
		})

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAllByOrgID-non-admin-jwt-%s", s.reqID))
			users, err := s.nonAdminUserClient.GetAllByOrgID(ctx, sysOrgID)
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(users), 1)
			found := false
			for _, u := range users {
				if u.ID == sysUserID {
					found = true
					assert.NotNil(t, u.ID)
					assert.Equal(t, sysUserID, u.ID)
					assert.Equal(t, sysOrgID, u.OrgID)
					assert.Equal(t, "System Admin", u.Name)
					assert.Equal(t, true, u.IsSystem)
					assert.Equal(t, true, u.IsAdmin)
					assert.Equal(t, true, u.IsActive)
					assert.Equal(t, int64(1), u.Version)
					assert.LessOrEqual(t, u.CreatedAt, u.UpdatedAt)
				}
			}
			assert.True(t, found)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("getAllByOrgID-invalid-jwt-%s", s.reqID))
			_, err := s.invJWTUserClient.GetAllByOrgID(ctx, sysOrgID)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-valid-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			assert.NotNil(t, u.ID)
			assert.Equal(t, int64(1), u.Version)
			assert.LessOrEqual(t, u.CreatedAt, u.UpdatedAt)
		})

		t.Run("MissingName", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-missing-name-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("MissingEmail", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-missing-email-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("DuplicateEmail", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-dup-email-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "john.ryan.bard@gmail.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 409, httpErr.StatusCode)
		})

		t.Run("SysUser", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-sys-user-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:     "Test-" + uuid.NewString(),
				Email:    "foo+" + uuid.NewString() + "@bar.com",
				OrgID:    s.testOrg.ID,
				IsSystem: true,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			assert.NotNil(t, u.ID)
			assert.False(t, u.IsSystem)
		})

		t.Run("OrgIDNotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-orgID-not-found-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: "will-not-find",
			})
			s.addUserToCleanup(u)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("SysOrgID", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-sysOrgID-not-found-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: sysOrgID,
			})
			s.addUserToCleanup(u)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-non-admin-jwt-%s", s.reqID))
			u, err := s.nonAdminUserClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("create-invalid-jwt-%s", s.reqID))
			u, err := s.invJWTUserClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-valid-setup-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-valid-%s", s.reqID))
			u.Name = u.Name + "-updated"
			u2, err := s.userClient.Save(ctx, u)
			s.addUserToCleanup(u2)
			assert.Nil(t, err)
			assert.Equal(t, u.Version+1, u2.Version)
			assert.Equal(t, u.CreatedAt, u2.CreatedAt)
			assert.Less(t, u.UpdatedAt, u2.UpdatedAt)
			assert.Less(t, u2.CreatedAt, u2.UpdatedAt)
		})

		t.Run("OptimisticLock", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-opt-lock-setup-1-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-opt-lock-setup-2-%s", s.reqID))
			u.Name = u.Name + "-updated"
			u2, err := s.userClient.Save(ctx, u)
			s.addUserToCleanup(u2)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-opt-lock-%s", s.reqID))
			u.Name = u.Name + "-updated-again"
			u3, err := s.userClient.Save(ctx, u)
			s.addUserToCleanup(u3)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 409, httpErr.StatusCode)
		})

		t.Run("MissingName", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-missing-name-setup-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-missing-name-%s", s.reqID))
			u.Name = ""
			u2, err := s.userClient.Save(ctx, u)
			s.addUserToCleanup(u2)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("MissingEmail", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-missing-email-setup-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-missing-email-%s", s.reqID))
			u.Email = ""
			u2, err := s.userClient.Save(ctx, u)
			s.addUserToCleanup(u2)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("NotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-not-found-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				ID:      "will-not-find",
				Name:    "Test-" + uuid.NewString(),
				Email:   "foo+" + uuid.NewString() + "@bar.com",
				OrgID:   s.testOrg.ID,
				Version: 1,
			})
			s.addUserToCleanup(u)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 404, httpErr.StatusCode)
		})

		t.Run("SysUser", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-sys-org-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				ID:      sysUserID,
				Name:    "Test-" + uuid.NewString(),
				Email:   "foo+" + uuid.NewString() + "@bar.com",
				OrgID:   s.testOrg.ID,
				Version: 1,
			})
			s.addUserToCleanup(u)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("OrgIDNotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-org-id-not-found-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			u.OrgID = "will-not-find"
			u2, err := s.userClient.Save(ctx, u)
			s.addUserToCleanup(u2)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("SysOrgID", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-sys-org-id-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			u.OrgID = sysOrgID
			u2, err := s.userClient.Save(ctx, u)
			s.addUserToCleanup(u2)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-non-admin-jwt-setup-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-non-admin-jwt-%s", s.reqID))
			u.Name = u.Name + "-updated"
			u2, err := s.nonAdminUserClient.Save(ctx, u)
			s.addUserToCleanup(u2)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-invalid-jwt-setup-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("update-invalid-jwt-%s", s.reqID))
			u.Name = u.Name + "-updated"
			u2, err := s.invJWTUserClient.Save(ctx, u)
			s.addUserToCleanup(u2)
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-valid-setup-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-valid-%s", s.reqID))
			err = s.userClient.Delete(ctx, user.DeleteUser{ID: u.ID, Version: u.Version})
			assert.Nil(t, err)
		})

		t.Run("Idempotent", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-idempotent-setup-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-idempotent-1-%s", s.reqID))
			err = s.userClient.Delete(ctx, user.DeleteUser{ID: u.ID, Version: u.Version})
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-idempotent-2-%s", s.reqID))
			err = s.userClient.Delete(ctx, user.DeleteUser{ID: u.ID, Version: u.Version})
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-idempotent-3-%s", s.reqID))
			err = s.userClient.Delete(ctx, user.DeleteUser{ID: u.ID, Version: u.Version})
			assert.Nil(t, err)
		})

		t.Run("OptimisticLock", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-opt-lock-setup-1-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-opt-lock-setup-2-%s", s.reqID))
			u.Name = u.Name + "-updated"
			u2, err := s.userClient.Save(ctx, u)
			s.addUserToCleanup(u2)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-opt-lock-%s", s.reqID))
			err = s.userClient.Delete(ctx, user.DeleteUser{ID: u.ID, Version: u.Version})
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 409, httpErr.StatusCode)
		})

		t.Run("MissingVersion", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-missing-version-setup-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-missing-version-%s", s.reqID))
			err = s.userClient.Delete(ctx, user.DeleteUser{ID: u.ID})
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 400, httpErr.StatusCode)
		})

		t.Run("NotFound", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-not-found-%s", s.reqID))
			err := s.userClient.Delete(ctx, user.DeleteUser{ID: "will-not-find", Version: 1})
			assert.Nil(t, err)
		})

		t.Run("SysUser", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-sys-org-%s", s.reqID))
			err := s.userClient.Delete(ctx, user.DeleteUser{ID: sysUserID, Version: 1})
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("NonAdminToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-non-admin-jwt-setup-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-non-admin-jwt-%s", s.reqID))
			err = s.nonAdminUserClient.Delete(ctx, user.DeleteUser{ID: u.ID, Version: u.Version})
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 403, httpErr.StatusCode)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-invalid-jwt-setup-%s", s.reqID))
			u, err := s.userClient.Save(ctx, user.User{
				Name:  "Test-" + uuid.NewString(),
				Email: "foo+" + uuid.NewString() + "@bar.com",
				OrgID: s.testOrg.ID,
			})
			s.addUserToCleanup(u)
			assert.Nil(t, err)
			ctx = context.WithValue(context.Background(), "reqID", fmt.Sprintf("delete-invalid-jwt-%s", s.reqID))
			err = s.invJWTUserClient.Delete(ctx, user.DeleteUser{ID: u.ID, Version: u.Version})
			assert.NotNil(t, err)
			var httpErr httpx.HTTPError
			assert.True(t, errors.As(err, &httpErr))
			assert.Equal(t, 401, httpErr.StatusCode)
		})
	})
}
