package kubernetes

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/events"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
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

func (l *logsStore) GetLogs(
	ctx context.Context,
	event brignext.Event,
	opts brignext.LogOptions,
) (brignext.LogEntryList, error) {
	logEntries := brignext.LogEntryList{}

	var podName string
	if opts.Job == "" {
		podName = fmt.Sprintf("worker-%s", event.ID)
	} else {
		podName = fmt.Sprintf("job-%s-%s", event.ID, strings.ToLower(opts.Job))
	}

	req := l.kubeClient.CoreV1().Pods(event.Kubernetes.Namespace).GetLogs(
		podName,
		&v1.PodLogOptions{
			Container:  opts.Container,
			Timestamps: true,
		},
	)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return logEntries, errors.Wrapf(
			err,
			"error opening log stream for pod %q in namespace %q",
			podName,
			event.Kubernetes.Namespace,
		)
	}
	defer podLogs.Close()

	logEntries.Items = []brignext.LogEntry{}

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
		logEntries.Items = append(logEntries.Items, logEntry)
	}

	return logEntries, nil
}

// TODO: Implement this
func (l *logsStore) StreamLogs(
	ctx context.Context,
	event brignext.Event,
	opts brignext.LogOptions,
) (<-chan brignext.LogEntry, error) {
	var podName string
	if opts.Job == "" {
		podName = fmt.Sprintf("worker-%s", event.ID)
	} else {
		podName = fmt.Sprintf("job-%s-%s", event.ID, strings.ToLower(opts.Job))
	}

	req := l.kubeClient.CoreV1().Pods(event.Kubernetes.Namespace).GetLogs(
		podName,
		&v1.PodLogOptions{
			Container:  opts.Container,
			Timestamps: true,
			Follow:     true,
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
	defer podLogs.Close()

	return nil, nil
}
