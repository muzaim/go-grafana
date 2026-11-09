package monitor

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidURL       = errors.New("invalid or malformed URL")
	ErrProtocolMismatch = errors.New("URL scheme does not match selected protocol")
	ErrInvalidTimeout   = errors.New("timeout must be less than check interval")
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, req CreateMonitorRequest) (*MonitorResponse, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*MonitorResponse, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) (*MonitorListResponse, error)
	Update(ctx context.Context, id, userID uuid.UUID, req UpdateMonitorRequest) (*MonitorResponse, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Toggle(ctx context.Context, id, userID uuid.UUID, isActive bool) (*MonitorResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) validateTarget(targetURL string, protocol Protocol, interval, timeout int) error {
	if timeout >= interval {
		return ErrInvalidTimeout
	}

	if protocol == ProtocolHTTP || protocol == ProtocolHTTPS {
		parsedURL, err := url.ParseRequestURI(targetURL)
		if err != nil || parsedURL.Host == "" {
			return ErrInvalidURL
		}

		scheme := strings.ToLower(parsedURL.Scheme)
		if (protocol == ProtocolHTTP && scheme != "http") || (protocol == ProtocolHTTPS && scheme != "https") {
			return ErrProtocolMismatch
		}
	} else if protocol == ProtocolTCP {
		// TCP target should be host:port
		if !strings.Contains(targetURL, ":") {
			return errors.New("TCP target must be in 'host:port' format")
		}
	}

	return nil
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, req CreateMonitorRequest) (*MonitorResponse, error) {
	if req.IntervalSec <= 0 {
		req.IntervalSec = 60
	}
	if req.TimeoutSec <= 0 {
		req.TimeoutSec = 10
	}
	if req.RetryThreshold <= 0 {
		req.RetryThreshold = 3
	}
	if req.ExpectedStatusCode <= 0 {
		req.ExpectedStatusCode = 200
	}
	if req.Method == "" {
		req.Method = "GET"
	}

	if err := s.validateTarget(req.URL, req.Protocol, req.IntervalSec, req.TimeoutSec); err != nil {
		return nil, err
	}

	m := &Monitor{
		UserID:             userID,
		Name:               req.Name,
		URL:                req.URL,
		Protocol:           req.Protocol,
		Method:             strings.ToUpper(req.Method),
		ExpectedStatusCode: req.ExpectedStatusCode,
		KeywordMatch:       req.KeywordMatch,
		IntervalSec:        req.IntervalSec,
		TimeoutSec:         req.TimeoutSec,
		RetryThreshold:     req.RetryThreshold,
		Status:             StatusPending,
		IsActive:           true,
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("failed to create monitor: %w", err)
	}

	resp := ToMonitorResponse(m)
	return &resp, nil
}

func (s *service) GetByID(ctx context.Context, id, userID uuid.UUID) (*MonitorResponse, error) {
	m, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	resp := ToMonitorResponse(m)
	return &resp, nil
}

func (s *service) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) (*MonitorListResponse, error) {
	monitors, total, err := s.repo.FindAllByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]MonitorResponse, len(monitors))
	for i := range monitors {
		responses[i] = ToMonitorResponse(&monitors[i])
	}

	return &MonitorListResponse{
		Monitors: responses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (s *service) Update(ctx context.Context, id, userID uuid.UUID, req UpdateMonitorRequest) (*MonitorResponse, error) {
	existing, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if req.IntervalSec <= 0 {
		req.IntervalSec = 60
	}
	if req.TimeoutSec <= 0 {
		req.TimeoutSec = 10
	}
	if req.RetryThreshold <= 0 {
		req.RetryThreshold = 3
	}
	if req.ExpectedStatusCode <= 0 {
		req.ExpectedStatusCode = 200
	}
	if req.Method == "" {
		req.Method = "GET"
	}

	if err := s.validateTarget(req.URL, req.Protocol, req.IntervalSec, req.TimeoutSec); err != nil {
		return nil, err
	}

	existing.Name = req.Name
	existing.URL = req.URL
	existing.Protocol = req.Protocol
	existing.Method = strings.ToUpper(req.Method)
	existing.ExpectedStatusCode = req.ExpectedStatusCode
	existing.KeywordMatch = req.KeywordMatch
	existing.IntervalSec = req.IntervalSec
	existing.TimeoutSec = req.TimeoutSec
	existing.RetryThreshold = req.RetryThreshold

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	resp := ToMonitorResponse(existing)
	return &resp, nil
}

func (s *service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *service) Toggle(ctx context.Context, id, userID uuid.UUID, isActive bool) (*MonitorResponse, error) {
	existing, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	existing.IsActive = isActive
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	resp := ToMonitorResponse(existing)
	return &resp, nil
}
