package service

import (
	"github.com/krancour/brignext/v2/internal/apiserver/scheduler"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
)

type Service interface {
	Events() EventsService
	Projects() ProjectsService
	ServiceAccounts() ServiceAccountsService
	Sessions() SessionsService
	Users() UsersService
}

type service struct {
	eventsService          EventsService
	projectsService        ProjectsService
	serviceAccountsService ServiceAccountsService
	sessionsService        SessionsService
	usersService           UsersService
}

func NewService(
	store storage.Store,
	scheduler scheduler.Scheduler,
	logStore storage.LogsStore,
) Service {
	return &service{
		eventsService:          NewEventsService(store, scheduler, logStore),
		projectsService:        NewProjectsService(store, scheduler),
		serviceAccountsService: NewServiceAccountsService(store.ServiceAccounts()),
		sessionsService:        NewSessionsService(store.Sessions()),
		usersService:           NewUsersService(store),
	}
}

func (s *service) Events() EventsService {
	return s.eventsService
}

func (s *service) Projects() ProjectsService {
	return s.projectsService
}

func (s *service) ServiceAccounts() ServiceAccountsService {
	return s.serviceAccountsService
}

func (s *service) Sessions() SessionsService {
	return s.sessionsService
}

func (s *service) Users() UsersService {
	return s.usersService
}
