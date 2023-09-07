package api

import (
	"context"
	"sort"

	"github.com/oklog/ulid/v2"
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

func (a *API) CreateGoal(ctx context.Context, req *goalsv1.CreateGoalRequest) (*goalsv1.CreateGoalResponse, error) {
	var site sitesv1.Site
	goal := &goalsv1.Goal{
		Id:        ulid.Make().String(),
		Type:      req.Type,
		Value:     req.Value,
		CreatedAt: timestamppb.New(core.Now(ctx)),
	}
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
		defer key.Release()
		err := txn.Get(key.Bytes(), px.Decode(&site), func() error {
			return status.Error(codes.NotFound, "site does not exist")
		})
		if err != nil {
			return err
		}
		for _, g := range site.Goals {
			if g.Type == req.Type && g.Value == req.Value {
				goal = g
				return nil
			}
		}
		if site.Goals == nil {
			site.Goals = make(map[string]*goalsv1.Goal)
		}
		site.Goals[goal.Id] = goal
		return txn.Set(key.Bytes(), px.Encode(&site))
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (a *API) GetGoal(ctx context.Context, req *goalsv1.GetGoalRequest) (*goalsv1.Goal, error) {
	var site sitesv1.Site
	var goal *goalsv1.Goal
	err := db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
		defer key.Release()
		err := txn.Get(key.Bytes(), px.Decode(&site), func() error {
			return status.Error(codes.NotFound, "site does not exist")
		})
		if err != nil {
			return err
		}
		if site.Goals != nil {
			goal = site.Goals[req.Id]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if goal == nil {
		return nil, status.Error(codes.NotFound, "goal does not exist")
	}
	return goal, nil
}

func (a *API) ListGoals(ctx context.Context, req *goalsv1.ListGoalsRequest) (*goalsv1.ListGoalsResponse, error) {
	var site sitesv1.Site
	var goals goalsv1.ListGoalsResponse
	err := db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
		defer key.Release()
		err := txn.Get(key.Bytes(), px.Decode(&site), func() error {
			return status.Error(codes.NotFound, "site does not exist")
		})
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
		defer key.Release()
		err := txn.Get(key.Bytes(), px.Decode(&site), func() error {
			return status.Error(codes.NotFound, "site does not exist")
		})
		if err != nil {
			return err
		}
		if site.Goals == nil || site.Goals[req.Id] == nil {
			return status.Error(codes.NotFound, "goal does not exist")
		}
		delete(site.Goals, req.Id)
		return txn.Set(key.Bytes(), px.Encode(&site))
	})
	if err != nil {
		return nil, err
	}
	return &goalsv1.DeleteGoalResponse{}, nil
}
