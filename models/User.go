package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID        bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	FirstName string        `json:"firstName"`
	LastName  string        `json:"lastName"`
	Email     string        `json:"email"`
	Password  string        `json:"password"`
}
