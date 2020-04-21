package brignext

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// nolint: lll
type Project struct {
	ID            string                   `json:"id" bson:"_id"`
	Description   string                   `json:"description" bson:"description"`
	Tags          ProjectTags              `json:"tags" bson:"tags"`
	WorkerConfigs map[string]WorkerConfig  `json:"workerConfigs" bson:"workerConfigs"`
	Kubernetes    *ProjectKubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Created       *time.Time               `json:"created,omitempty" bson:"created"`
}

type ProjectTags map[string]string

func (p *Project) GetWorkers(event Event) map[string]Worker {
	workers := map[string]Worker{}
	for workerName, workerConfig := range p.WorkerConfigs {
		if workerConfig.Matches(event.Source, event.Type) {
			worker := Worker{
				Container:            workerConfig.Container,
				WorkspaceSize:        workerConfig.WorkspaceSize,
				Git:                  workerConfig.Git,
				JobsConfig:           workerConfig.JobsConfig,
				LogLevel:             workerConfig.LogLevel,
				ConfigFilesDirectory: workerConfig.ConfigFilesDirectory,
				Jobs:                 map[string]Job{},
				Status: WorkerStatus{
					Phase: WorkerPhasePending,
				},
			}
			if worker.Container.Image == "" {
				worker.Container.Image = "krancour/brignext-worker:latest"
			}
			if worker.Container.ImagePullPolicy == "" {
				worker.Container.ImagePullPolicy = "IfNotPresent"
			}
			if worker.WorkspaceSize == "" {
				worker.WorkspaceSize = "10Gi"
			}

			// VCS details from the event override project-level details
			if event.Git.CloneURL != "" {
				worker.Git.CloneURL = event.Git.CloneURL
			}
			if event.Git.Commit != "" {
				worker.Git.Commit = event.Git.Commit
			}
			if event.Git.Ref != "" {
				worker.Git.Ref = event.Git.Ref
			}

			if worker.Git.CloneURL != "" &&
				worker.Git.Commit == "" &&
				worker.Git.Ref == "" {
				worker.Git.Ref = "master"
			}

			if worker.ConfigFilesDirectory == "" {
				worker.ConfigFilesDirectory = ".brigade"
			}

			workers[workerName] = worker
		}
	}
	return workers
}

// UnmarshalBSON implements custom BSON marshaling for the Project type. This
// does little more than guarantees that the Tags field isn't nil so that custom
// marshaling of the ProjectTags (which is more involved) can succeed.
func (p *Project) UnmarshalBSON(bytes []byte) error {
	if p.Tags == nil {
		p.Tags = ProjectTags{}
	}
	type ProjectAlias Project
	return bson.Unmarshal(
		bytes,
		&struct {
			*ProjectAlias `bson:",inline"`
		}{
			ProjectAlias: (*ProjectAlias)(p),
		},
	)
}

// MarshalBSONValue implements custom BSON marshaling for the ProjectTags type.
// ProjectTags is, essentially, a map[string]string, but when marshaled to BSON,
// it must be represented as follows because Mongo can index this more easily,
// making for faster queries:
//
// [
//   { "key": "key0", "value": "value0" },
//   { "key": "key1", "value": "value1" },
//   ...
//   { "key": "keyN", "value": "valueN" }
// ]
func (p ProjectTags) MarshalBSONValue() (bsontype.Type, []byte, error) {
	ms := make([]bson.M, len(p))
	var i int
	for k, v := range p {
		ms[i] = bson.M{
			"key":   k,
			"value": v,
		}
		i++
	}
	return bson.MarshalValue(ms)
}

// UnmarshalBSONValue implements custom BSON unmarshaling for the ProjectTags
// type. ProjectTags is, essentially, a map[string]string, but when marshaled to
// BSON, it is represented as follows because Mongo can index this more easily,
// making for faster queries:
//
// [
//   { "key": "key0", "value": "value0" },
//   { "key": "key1", "value": "value1" },
//   ...
//   { "key": "keyN", "value": "valueN" }
// ]
func (p ProjectTags) UnmarshalBSONValue(_ bsontype.Type, bytes []byte) error {
	tags := bson.M{}
	if err := bson.Unmarshal(bytes, &tags); err != nil {
		return err
	}
	for _, tag := range tags {
		t := tag.(bson.M)
		k := t["key"].(string)
		v := t["value"].(string)
		p[k] = v
	}
	return nil
}
