package server

import (
	"context"
	"encoding/json"
	"fmt"
	raft_internal "kv-storage/internal/raft"
	"kv-storage/internal/store"
	"kv-storage/proto"
	"time"

	"kv-storage/internal/metrics"

	"github.com/hashicorp/raft"
	"github.com/prometheus/client_golang/prometheus"
)

type Server struct {
	proto.UnimplementedKVServiceServer
	Store *store.Store
	Raft  *raft.Raft
}

func NewServer(store *store.Store, r *raft.Raft) *Server {
	return &Server{Store: store, Raft: r}
}

func (s *Server) Set(ctx context.Context, req *proto.SetRequest) (*proto.SetResponse, error) {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("set"))
	defer timer.ObserveDuration()
	cmd := raft_internal.Command{
		Op:    "SET",
		Key:   req.Key,
		Value: req.Value,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	applyFuture := s.Raft.Apply(data, 1*time.Second)

	if err := applyFuture.Error(); err != nil {
		metrics.RequestsTotal.WithLabelValues("set", "error").Inc()
		return &proto.SetResponse{Success: false}, err
	}

	fsmResponse := applyFuture.Response()
	if fsmResponse != nil {
		if err, ok := fsmResponse.(error); ok {
			metrics.RequestsTotal.WithLabelValues("set", "error").Inc()
			return &proto.SetResponse{Success: false}, err
		}
	}

	metrics.RequestsTotal.WithLabelValues("set", "success").Inc()
	return &proto.SetResponse{Success: true}, nil
}

func (s *Server) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("get"))
	defer timer.ObserveDuration()
	entry, err := s.Store.Get(req.Key)
	if err != nil {
		metrics.RequestsTotal.WithLabelValues("get", "error").Inc()
		return nil, err
	}
	metrics.RequestsTotal.WithLabelValues("get", "success").Inc()
	return &proto.GetResponse{Value: entry.Value}, nil
}

func (s *Server) Join(ctx context.Context, req *proto.JoinRequest) (*proto.JoinResponse, error) {
	if s.Raft.State() != raft.Leader {
		return nil, fmt.Errorf("not the leader")
	}

	future := s.Raft.AddVoter(raft.ServerID(req.NodeID), raft.ServerAddress(req.RaftAddress), 0, 0)
	if err := future.Error(); err != nil {
		return nil, err
	}
	return &proto.JoinResponse{}, nil
}
