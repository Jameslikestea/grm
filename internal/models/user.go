package models

import (
	"time"

	"github.com/go-git/go-git/v5/plumbing"
)

type Model interface{}

var _ Model = User{}
var _ Model = UserSession{}

// User is the model of connected users from 3rd party providers
// such as github
type User struct {
	UID  string
	Hash plumbing.Hash
}

// UserSession is a model of a connected users session.
type UserSession struct {
	Hash    plumbing.Hash
	User    User
	Expires time.Time
}
