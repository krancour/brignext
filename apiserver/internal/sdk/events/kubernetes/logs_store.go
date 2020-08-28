package kubernetes

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk/events"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type logsStore struct {
	kubeClient *kubernetes.Clientset
}

func NewLogsStore(kubeClient *kubernetes.Clientset) events.LogsStore {
	return &logsStore{
		kubeClient: kubeClient,
	}
}

func (l *logsStore) StreamLogs(
	ctx context.Context,
	event brignext.Event,
	selector brignext.LogsSelector,
	opts brignext.LogStreamOptions,
) (<-chan brignext.LogEntry, error) {
	var podName string
	if selector.Job == "" {
		podName = fmt.Sprintf("worker-%s", event.ID)
	} else {
		podName = fmt.Sprintf("job-%s-%s", event.ID, strings.ToLower(selector.Job))
	}

	req := l.kubeClient.CoreV1().Pods(event.Kubernetes.Namespace).GetLogs(
		podName,
		&v1.PodLogOptions{
			Container:  selector.Container,
			Timestamps: true,
		},
	)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error opening log stream for pod %q in namespace %q",
			podName,
			event.Kubernetes.Namespace,
		)
	}

	logEntryCh := make(chan brignext.LogEntry)

	go func() {
		defer podLogs.Close()
		defer close(logEntryCh)
		buffer := bufio.NewReader(podLogs)
		for {
			logEntry := brignext.LogEntry{}
			logLine, err := buffer.ReadString('\n')
			if err == io.EOF {
				break
			}
			// The last character should be a newline that we don't want, so let's
			// remove that
			logLine = logLine[:len(logLine)-1]
			logLineParts := strings.SplitN(logLine, " ", 2)
			if len(logLineParts) == 2 {
				timeStr := logLineParts[0]
				t, err := time.Parse(time.RFC3339, timeStr)
				if err == nil {
					logEntry.Time = &t
				}
				logEntry.Message = logLineParts[1]
			} else {
				logEntry.Message = logLine
			}
			select {
			case logEntryCh <- logEntry:
			case <-ctx.Done():
				return
			}
		}
		podLogs.Close()
		if opts.Follow {
			// If following, we let this goroutine hang until the context times out or
			// is canceled. Why? When logs are followed from COLD storage (e.g.
			// MongoDB) we never know whether the log aggregator has stored all the
			// logs we're trying to stream, so we don't disconnect since there's a
			// possibility there is more coming. We leave it up to the client to
			// decide to disconnect. For consistency, we're leaving it up to the
			// client to disconnect here as well. We can revisit this if we can make
			// the COLD log storage smarter about knowing when it has reached the end
			// of a stream, in which case both warm and cold storage could both
			// disconnect when the end of a stream is reached and they would still be
			// consistent with one another.
			<-ctx.Done()
		}
	}()

	return logEntryCh, nil
}
