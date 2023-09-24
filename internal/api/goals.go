package api

import (
	"context"
	"sort"

	goalsv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/goals/v1"
	sitesv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/px"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ goalsv1.GoalsServer = (*API)(nil)

var goalE404 = status.Error(codes.NotFound, "goal does not exist")

func GoalE404() error {
	return goalE404
}

func (a *API) CreateGoal(ctx context.Context, req *goalsv1.CreateGoalRequest) (*goalsv1.CreateGoalResponse, error) {
	key := keys.Site(req.Domain)

	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		var site sitesv1.Site
		err := txn.Get(key, px.Decode(&site), Sites404)
		if err != nil {
			return err
		}
		if site.Goals != nil && site.Goals[req.Name] != nil {
			return status.Error(codes.AlreadyExists, "goal already exists")
		}
		if site.Goals == nil {
			site.Goals = make(map[string]*goalsv1.Goal)
		}
		site.Goals[req.Name] = &goalsv1.Goal{
			Name:      req.Name,
			Type:      req.Type,
			Value:     req.Value,
			CreatedAt: timestamppb.New(core.Now(ctx)),
		}
		return txn.Set(key, px.Encode(&site))
	})
	if err != nil {
		return nil, err
	}
	return &goalsv1.CreateGoalResponse{}, nil
}

func (a *API) GetGoal(ctx context.Context, req *goalsv1.GetGoalRequest) (*goalsv1.Goal, error) {
	var goal *goalsv1.Goal
	key := keys.Site(req.Domain)
	err := db.Get(ctx).Txn(false, func(txn db.Txn) error {
		var site sitesv1.Site
		err := txn.Get(key, px.Decode(&site), Sites404)
		if err != nil {
			return err
		}
		if site.Goals != nil {
			goal = site.Goals[req.Name]
		}
		if goal == nil {
			return goalE404
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return goal, nil
}

func (a *API) ListGoals(ctx context.Context, req *goalsv1.ListGoalsRequest) (*goalsv1.ListGoalsResponse, error) {
	var site sitesv1.Site
	var goals goalsv1.ListGoalsResponse
	err := db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
		err := txn.Get(key, px.Decode(&site), Sites404)
		if err != nil {
			return err
		}
		for _, g := range site.Goals {
			goals.Goals = append(goals.Goals, g)
		}
		sort.Slice(goals.Goals, func(i, j int) bool {
			return goals.Goals[i].GetCreatedAt().AsTime().Before(
				goals.Goals[j].CreatedAt.AsTime(),
			)
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &goals, nil
}

func (a *API) DeleteGoal(ctx context.Context, req *goalsv1.DeleteGoalRequest) (*goalsv1.DeleteGoalResponse, error) {
	var site sitesv1.Site
	err := db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
		err := txn.Get(key, px.Decode(&site), Sites404)
		if err != nil {
			return err
		}
		if site.Goals == nil || site.Goals[req.Name] == nil {
			return goalE404
		}
		delete(site.Goals, req.Name)
		return txn.Set(key, px.Encode(&site))
	})
	if err != nil {
		return nil, err
	}
	return &goalsv1.DeleteGoalResponse{}, nil
}
