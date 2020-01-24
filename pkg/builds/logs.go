package builds

import (
	"github.com/pkg/errors"
)

func (b *buildsServer) StreamBuildLogs(
	req *StreamBuildLogsRequest,
	stream Builds_StreamBuildLogsServer,
) error {
	// TODO: We should do some kind of validation!

	logEntryCh, err := b.logStore.StreamWorkerLogs(stream.Context(), req.Id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving logs for build %q", req.Id)
	}

	for logEntry := range logEntryCh {
		err = stream.Send(
			&LogEntry{
				Time:    logEntry.Time.UnixNano(),
				Message: logEntry.Message,
			},
		)
		if err != nil {
			return errors.Wrap(err, "error sending log entry")
		}
	}

	return nil
}
