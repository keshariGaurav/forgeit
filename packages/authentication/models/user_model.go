package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id       primitive.ObjectID `json:"id,omitempty"`
	Email    string             `json:"email,omitempty" validate:"required,email"`
	Name     string             `json:"name,omitempty" validate:"required"`
	Location string             `json:"location,omitempty" validate:"required"`
	Title    string             `json:"title,omitempty" validate:"required"`
	Address  string             `json:"address,omitempty" validate:"required"`
	LinkedIn string             `json:"linkedin,omitempty" validate:"required,url"`
	Twitter  string             `json:"twitter,omitempty" validate:"url"`
	DOB      string             `json:"dob,omitempty" validate:"required,datetime=2006-01-02"`
	Resume   string             `json:"resume,omitempty"`
	Username string 						`json:"username,omitempty" validate:"required"`
}
