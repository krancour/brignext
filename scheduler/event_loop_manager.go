package main

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext/v2/sdk/api"
	"github.com/krancour/brignext/v2/sdk/meta"
)

func (s *scheduler) manageProjectEventLoops(ctx context.Context) {
	// Maintain a map of functions for canceling the event loops for each known
	// project
	eventLoopCancelFns := map[string]func(){}

	// TODO: Is it ok that this is hardcoded?
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		projects, err :=
			s.apiClient.Projects().List(ctx, api.ProjectsSelector{}, meta.ListOptions{})
		if err != nil {
			select {
			case s.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		// Build a set of current projects. This makes it a little faster and easier
		// to search for projects later in this algorithm.
		currentProjects := map[string]struct{}{}
		for _, project := range projects.Items {
			currentProjects[project.ID] = struct{}{}
		}

		// Reconcile differences between projects we knew about already and the
		// current set of projects...

		// 1. Stop event loops for projects that have been deleted
		for projectID, cancelFn := range eventLoopCancelFns {
			if _, stillExists := currentProjects[projectID]; !stillExists {
				log.Printf("DEBUG: stopping event loop for project %q", projectID)
				cancelFn()
				// Surprisingly, Go lets us delete from a map we are currently iterating
				// over. How convenient.
				delete(eventLoopCancelFns, projectID)
			}
		}

		// 2. Start event loop for any projects that have been added
		for projectID := range currentProjects {
			if _, known := eventLoopCancelFns[projectID]; !known {
				eventLoopCtx, eventLoopCtxCancelFn := context.WithCancel(ctx)
				eventLoopCancelFns[projectID] = eventLoopCtxCancelFn
				log.Printf("DEBUG: starting event loop for project %q", projectID)
				go s.runEventLoop(eventLoopCtx, projectID)
			}
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}

}
