package service

import (
	"github.com/krancour/brignext/v2/internal/apiserver/scheduler"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
)

type Service interface {
	Sessions() SessionsService
	Users() UsersService
	ServiceAccounts() ServiceAccountsService
	Projects() ProjectsService
	Events() EventsService
}

type service struct {
	sessionsService        SessionsService
	usersService           UsersService
	serviceAccountsService ServiceAccountsService
	projectsService        ProjectsService
	eventsService          EventsService
}

func NewService(
	store storage.Store,
	scheduler scheduler.Scheduler,
	logStore storage.LogsStore,
) Service {
	return &service{
		sessionsService:        NewSessionsService(store),
		usersService:           NewUsersService(store),
		serviceAccountsService: NewServiceAccountsService(store),
		projectsService:        NewProjectsService(store, scheduler),
		eventsService:          NewEventsService(store, scheduler, logStore),
	}
}

func (s *service) Sessions() SessionsService {
	return s.sessionsService
}

func (s *service) Users() UsersService {
	return s.usersService
}

func (s *service) ServiceAccounts() ServiceAccountsService {
	return s.serviceAccountsService
}

func (s *service) Projects() ProjectsService {
	return s.projectsService
}

func (s *service) Events() EventsService {
	return s.eventsService
}
