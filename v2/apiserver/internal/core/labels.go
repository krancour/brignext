package core

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// Labels is a map of key/value pairs utilized by Events in describing
// themselves and by EventSubscriptions in describing Events of interest to a
// Project.
type Labels map[string]string

// MarshalBSONValue implements custom BSON marshaling for the Labels type.
// Labels is, essentially, a map[string]string, but when marshaled to BSON,
// it must be represented as follows because Mongo can index this more easily,
// making for faster queries:
//
// [
//   { "key": "key0", "value": "value0" },
//   { "key": "key1", "value": "value1" },
//   ...
//   { "key": "keyN", "value": "valueN" }
// ]
func (l Labels) MarshalBSONValue() (bsontype.Type, []byte, error) {
	ms := make([]bson.M, len(l))
	var i int
	for k, v := range l {
		ms[i] = bson.M{
			"key":   k,
			"value": v,
		}
		i++
	}
	return bson.MarshalValue(ms)
}

// UnmarshalBSONValue implements custom BSON unmarshaling for the Labels
// type. Labels is, essentially, a map[string]string, but when marshaled to
// BSON, it is represented as follows because Mongo can index this more easily,
// making for faster queries:
//
// [
//   { "key": "key0", "value": "value0" },
//   { "key": "key1", "value": "value1" },
//   ...
//   { "key": "keyN", "value": "valueN" }
// ]
func (l Labels) UnmarshalBSONValue(_ bsontype.Type, bytes []byte) error {
	labels := bson.M{}
	if err := bson.Unmarshal(bytes, &labels); err != nil {
		return err
	}
	for _, label := range labels {
		m := label.(bson.M)
		k := m["key"].(string)
		v := m["value"].(string)
		l[k] = v
	}
	return nil
}
