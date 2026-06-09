package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Game struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RawgID        int                `bson:"rawg_id" json:"rawg_id"`
	Title         string             `bson:"title" json:"title"`
	Genre         string             `bson:"genre,omitempty" json:"genre,omitempty"`
	Platform      string             `bson:"platform,omitempty" json:"platform,omitempty"`
	CoverURL      string             `bson:"cover_url,omitempty" json:"cover_url,omitempty"`
	PersonalNote  string             `bson:"personal_note,omitempty" json:"personal_note,omitempty"`
	PersonalScore *int               `bson:"personal_score,omitempty" json:"personal_score,omitempty"`
	Status        string             `bson:"status,omitempty" json:"status,omitempty"`
	AddedAt       time.Time          `bson:"added_at,omitempty" json:"added_at,omitempty"`
}
