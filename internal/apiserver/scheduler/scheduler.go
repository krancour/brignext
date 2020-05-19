package scheduler

import (
	"github.com/go-redis/redis" // nolint: lll
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
}

type scheduler struct {
	projectsScheduler ProjectsScheduler
	eventsScheduler   EventsScheduler
}

func NewScheduler(
	redisClient *redis.Client,
	kubeClient *kubernetes.Clientset,
) Scheduler {
	return &scheduler{
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
