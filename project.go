package brignext

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// nolint: lll
type Project struct {
	ID          string                  `json:"id" bson:"_id,omitempty"`
	Description string                  `json:"description,omitempty" bson:"description,omitempty"`
	Tags        ProjectTags             `json:"tags,omitempty" bson:"tags,omitempty"`
	Workers     map[string]WorkerConfig `json:"workers,omitempty" bson:"workers,omitempty"`
	Namespace   string                  `json:"namespace,omitempty" bson:"namespace,omitempty"`
	// TODO: Secrets should be broken out into their own thing and shouldn't be a project field
	Secrets map[string]string `json:"secrets,omitempty" bson:"-"`
	Created *time.Time        `json:"created,omitempty" bson:"created,omitempty"`
	// DefaultScript     string     `json:"defaultScript,omitempty" bson:"defaultScript,omitempty"`
	// DefaultScriptName string     `json:"defaultScriptName,omitempty" bson:"defaultScriptName,omitempty"`
	// DefaultConfig     string     `json:"defaultConfig,omitempty" bson:"defaultConfig,omitempty"`
	// DefaultConfigName string     `json:"defaultConfigName,omitempty" bson:"defaultConfigName,omitempty"`
	// Github       Github `json:"github,omitempty" bson:"github,omitempty"`
	// InitGitSubmodules    bool              `json:"initGitSubmodules,omitempty" bson:"initGitSubmodules,omitempty"`
	// ImagePullSecrets     string            `json:"imagePullSecrets,omitempty" bson:"imagePullSecrets,omitempty"`
	// BrigadejsPath        string            `json:"brigadejsPath,omitempty" bson:"brigadejsPath,omitempty"`
	// BrigadeConfigPath    string            `json:"brigadeConfigPath,omitempty" bson:"brigadeConfigPath,omitempty"`
}

type ProjectTags map[string]string

func (p *Project) GetWorkers(
	eventProvider string,
	eventType string,
) map[string]Worker {
	workers := map[string]Worker{}
	for workerName, workerConfig := range p.Workers {
		if workerConfig.Matches(eventProvider, eventType) {
			workers[workerName] = Worker{
				InitContainer: workerConfig.InitContainer,
				Container:     workerConfig.Container,
				WorkspaceSize: workerConfig.WorkspaceSize,
				Git:           workerConfig.Git,
				Kubernetes:    workerConfig.Kubernetes,
				Jobs:          workerConfig.Jobs,
				Status:        WorkerStatusPending,
			}
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
