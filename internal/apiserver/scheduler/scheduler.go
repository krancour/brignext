package scheduler

import (
	"context"

	"github.com/go-redis/redis" // nolint: lll
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

const (
	componentLabel = "brignext.io/component"
	projectLabel   = "brignext.io/project"
	eventLabel     = "brignext.io/event"
)

type Scheduler interface {
	Projects() ProjectsScheduler
	Events() EventsScheduler

	CheckHealth(context.Context) error
}

type scheduler struct {
	redisClient       *redis.Client
	projectsScheduler ProjectsScheduler
	eventsScheduler   EventsScheduler
}

func NewScheduler(
	redisClient *redis.Client,
	kubeClient *kubernetes.Clientset,
) Scheduler {
	return &scheduler{
		redisClient:       redisClient,
		projectsScheduler: NewProjectsScheduler(kubeClient),
		eventsScheduler:   NewEventsScheduler(redisClient, kubeClient),
	}
}

func (s *scheduler) Projects() ProjectsScheduler {
	return s.projectsScheduler
}

func (s *scheduler) Events() EventsScheduler {
	return s.eventsScheduler
}

func (s *scheduler) CheckHealth(context.Context) error {
	if err := s.redisClient.Ping().Err(); err != nil {
		return errors.Wrap(err, "error pinging redis")
	}
	return nil
}
