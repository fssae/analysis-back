package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type VideoMessage struct {
	Id  primitive.ObjectID `json:"_id" bson:"_id"`
	Url string             `bson:"url" json:"url"`
}
