package vo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ID struct {
	Value primitive.ObjectID `json:"value" bson:"_id,omitempty"`
}

func NewID(oid primitive.ObjectID) ID {
	return ID{Value: oid}
}

func (id *ID) Hex() string {
	return id.Value.Hex()
}
