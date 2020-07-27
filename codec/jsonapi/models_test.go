package jsonapi

import (
	"time"
)

//go:generate neuron-generator models --format=goimports --single-file --exclude=Node .

// UnmarshalModel basic model used to tests in unmarshal
type UnmarshalModel struct {
	ID          string     `neuron:"type=primary"`
	PtrString   *string    `neuron:"type=attr"`
	PtrTime     *time.Time `neuron:"type=attr"`
	StringSlice []string   `neuron:"type=attr"`
}

// User defines a testing model with some private field attributes and many2many relationship
type User struct {
	privateField int
	ID           int    `neuron:"type=primary"`
	Lang         string `neuron:"type=attr;name=lang;flags=lang"`
	Name         string `neuron:"type=attr;name=name"`
	Pets         []*Pet `neuron:"type=relation;name=pets;many2many=UserPets"`
}

// Pet defines a testing model that contains attribute and a relation many2many
type Pet struct {
	ID     int     `neuron:"type=primary"`
	Name   string  `neuron:"type=attr;name=name"`
	Owners []*User `neuron:"type=relation;name=owners;many2many=UserPets"`
}

// UserPets is the join model for the User and Pets many2many relationship
type UserPets struct {
	ID     int `neuron:"type=primary"`
	PetID  int `neuron:"type=foreign"`
	UserID int `neuron:"type=foreign"`
}

/* HasMany Example */

// Blog defines a test model containing multiple relationships to the Post
type Blog struct {
	ID            int       `neuron:"type=primary"`
	Title         string    `neuron:"type=attr;name=title"`
	Posts         []*Post   `neuron:"type=relation;name=posts"`
	CurrentPost   *Post     `neuron:"type=relation;name=current_post"`
	CurrentPostID uint64    `neuron:"type=foreign;name=current_post_id"`
	CreatedAt     time.Time `neuron:"type=attr;name=created_at;flags=iso8601"`
	ViewCount     int       `neuron:"type=attr;name=view_count;flags=omitempty"`
}

// Post is a test model that is related to the blog and comments
type Post struct {
	ID            uint64     `neuron:"type=primary"`
	BlogID        int        `neuron:"type=foreign;name=blog_id"`
	Title         string     `neuron:"type=attr;name=title"`
	Body          string     `neuron:"type=attr;name=body"`
	Comments      []*Comment `neuron:"type=relation;name=comments;foreign=PostID"`
	LatestComment *Comment   `neuron:"type=relation;name=latest_comment;foreign=PostID"`
}

// Comment is a test model related to the post
type Comment struct {
	ID     int    `neuron:"type=primary"`
	PostID uint64 `neuron:"type=foreign;name=post_id"`
	Body   string `neuron:"type=attr;name=body"`
}

// ArrModel is testing model with array attribute.
type ArrModel struct {
	ID  int        `neuron:"type=primary"`
	Arr [2]float64 `neuron:"type=attr"`
}

// ModelPtrTime is testing model with time pointer.
type ModelPtrTime struct {
	ID   int        `neuron:"type=primary"`
	Time *time.Time `neuron:"type=attr"`
}

// ModelTime is testing model with time attribute.
type ModelTime struct {
	ID   int       `neuron:"type=primary"`
	Time time.Time `neuron:"type=attr"`
}

// HiddenModel is the neuron model with hidden fields.
type HiddenModel struct {
	ID          int    `neuron:"type=primary;flags=hidden"`
	Visible     string `neuron:"type=attr"`
	HiddenField string `neuron:"type=attr;flags=hidden"`
}

// CustomTagModel is the model with field that contains custom codec tag.
type CustomTagModel struct {
	ID                int
	VisibleCustomName string `codec:"visible"`
	HiddenField       bool   `codec:"-"`
	OmitEmptyField    string `codec:"omitempty"`
	CustomOmitEmpty   string `codec:"custom,omitempty"`
}
