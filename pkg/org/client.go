package org

import (
	"context"
	"fmt"
	"github.com/RyanBard/gin-ex/internal/apiclient"
)

type Config struct {
	BaseURL string
}

type orgClient struct {
	cfg Config
	ac  *apiclient.Client
}

func NewClient(cfg Config, ac *apiclient.Client) *orgClient {
	return &orgClient{
		cfg: cfg,
		ac:  ac,
	}
}

func (oc *orgClient) GetByID(ctx context.Context, id string) (o Org, err error) {
	path := fmt.Sprintf("%s/api/orgs/:id", oc.cfg.BaseURL)
	pathParams := map[string]string{
		"id": id,
	}
	queryParams := map[string][]string{}
	err = oc.ac.Get(ctx, path, pathParams, queryParams, &o)
	return o, err
}

func (oc *orgClient) GetAll(ctx context.Context) (o []Org, err error) {
	path := fmt.Sprintf("%s/api/orgs", oc.cfg.BaseURL)
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	err = oc.ac.Get(ctx, path, pathParams, queryParams, &o)
	return o, err
}

func (oc *orgClient) SearchByName(ctx context.Context, name string) (o []Org, err error) {
	path := fmt.Sprintf("%s/api/orgs", oc.cfg.BaseURL)
	pathParams := map[string]string{}
	queryParams := map[string][]string{
		"name": {name},
	}
	err = oc.ac.Get(ctx, path, pathParams, queryParams, &o)
	return o, err
}

func (oc *orgClient) Save(ctx context.Context, input Org) (o Org, err error) {
	queryParams := map[string][]string{}
	if input.ID == "" {
		path := fmt.Sprintf("%s/api/orgs", oc.cfg.BaseURL)
		pathParams := map[string]string{}
		err = oc.ac.Post(ctx, path, pathParams, queryParams, input, &o)
	} else {
		path := fmt.Sprintf("%s/api/orgs/:id", oc.cfg.BaseURL)
		pathParams := map[string]string{
			"id": input.ID,
		}
		err = oc.ac.Put(ctx, path, pathParams, queryParams, input, &o)
	}
	return o, err
}

func (oc *orgClient) Delete(ctx context.Context, input DeleteOrg) (err error) {
	path := fmt.Sprintf("%s/api/orgs/:id", oc.cfg.BaseURL)
	pathParams := map[string]string{
		"id": input.ID,
	}
	queryParams := map[string][]string{}
	return oc.ac.Delete(ctx, path, pathParams, queryParams, input, nil)
}
