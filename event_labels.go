package brignext

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type EventLabels map[string]string

// MarshalBSONValue implements custom BSON marshaling for the EventLabels type.
// EventLabels is, essentially, a map[string]string, but when marshaled to BSON,
// it must be represented as follows because Mongo can index this more easily,
// making for faster queries:
//
// [
//   { "key": "key0", "value": "value0" },
//   { "key": "key1", "value": "value1" },
//   ...
//   { "key": "keyN", "value": "valueN" }
// ]
func (e EventLabels) MarshalBSONValue() (bsontype.Type, []byte, error) {
	ms := make([]bson.M, len(e))
	var i int
	for k, v := range e {
		ms[i] = bson.M{
			"key":   k,
			"value": v,
		}
		i++
	}
	return bson.MarshalValue(ms)
}

// UnmarshalBSONValue implements custom BSON unmarshaling for the EventLabels
// type. EventLabels is, essentially, a map[string]string, but when marshaled to
// BSON, it is represented as follows because Mongo can index this more easily,
// making for faster queries:
//
// [
//   { "key": "key0", "value": "value0" },
//   { "key": "key1", "value": "value1" },
//   ...
//   { "key": "keyN", "value": "valueN" }
// ]
func (e EventLabels) UnmarshalBSONValue(_ bsontype.Type, bytes []byte) error {
	labels := bson.M{}
	if err := bson.Unmarshal(bytes, &labels); err != nil {
		return err
	}
	for _, label := range labels {
		l := label.(bson.M)
		k := l["key"].(string)
		v := l["value"].(string)
		e[k] = v
	}
	return nil
}
