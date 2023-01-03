package user

import (
	"context"
	"fmt"
	"github.com/RyanBard/gin-ex/internal/apiclient"
)

type Config struct {
	BaseURL string
}

type userClient struct {
	cfg Config
	ac  *apiclient.Client
}

func NewClient(cfg Config, ac *apiclient.Client) *userClient {
	return &userClient{
		cfg: cfg,
		ac:  ac,
	}
}

func (uc *userClient) GetByID(ctx context.Context, id string) (u User, err error) {
	path := fmt.Sprintf("%s/api/users/:id", uc.cfg.BaseURL)
	pathParams := map[string]string{
		"id": id,
	}
	queryParams := map[string][]string{}
	err = uc.ac.Get(ctx, path, pathParams, queryParams, &u)
	return u, err
}

func (uc *userClient) GetAll(ctx context.Context) (u []User, err error) {
	path := fmt.Sprintf("%s/api/users", uc.cfg.BaseURL)
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	err = uc.ac.Get(ctx, path, pathParams, queryParams, &u)
	return u, err
}

func (uc *userClient) GetAllByOrgID(ctx context.Context, orgID string) (u []User, err error) {
	path := fmt.Sprintf("%s/api/orgs/:orgID/users", uc.cfg.BaseURL)
	pathParams := map[string]string{
		"orgID": orgID,
	}
	queryParams := map[string][]string{}
	err = uc.ac.Get(ctx, path, pathParams, queryParams, &u)
	return u, err
}

func (uc *userClient) Save(ctx context.Context, input User) (u User, err error) {
	queryParams := map[string][]string{}
	if input.ID == "" {
		path := fmt.Sprintf("%s/api/users", uc.cfg.BaseURL)
		pathParams := map[string]string{}
		err = uc.ac.Post(ctx, path, pathParams, queryParams, input, &u)
	} else {
		path := fmt.Sprintf("%s/api/users/:id", uc.cfg.BaseURL)
		pathParams := map[string]string{
			"id": input.ID,
		}
		err = uc.ac.Put(ctx, path, pathParams, queryParams, input, &u)
	}
	return u, err
}

func (uc *userClient) Delete(ctx context.Context, input DeleteUser) (err error) {
	path := fmt.Sprintf("%s/api/users/:id", uc.cfg.BaseURL)
	pathParams := map[string]string{
		"id": input.ID,
	}
	queryParams := map[string][]string{}
	return uc.ac.Delete(ctx, path, pathParams, queryParams, input, nil)
}
