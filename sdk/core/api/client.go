package api

type Client interface {
	// Events returns a specialized client for Event management.
	Events() EventsClient
	// Projects returns a specialized client for Project management.
	Projects() ProjectsClient
}

type client struct {
	// eventsClient is a specialized client for Event management.
	eventsClient EventsClient
	// projectsClient is a specialized client for Project management.
	projectsClient ProjectsClient
}

func NewClient(apiAddress, apiToken string, allowInsecure bool) Client {
	return &client{
		eventsClient:   NewEventsClient(apiAddress, apiToken, allowInsecure),
		projectsClient: NewProjectsClient(apiAddress, apiToken, allowInsecure),
	}
}

func (c *client) Events() EventsClient {
	return c.eventsClient
}

func (c *client) Projects() ProjectsClient {
	return c.projectsClient
}
