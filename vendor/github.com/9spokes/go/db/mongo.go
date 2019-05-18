package db

import (
	"github.com/globalsign/mgo"
)

// MongoDB is a struct representing a MongoDB instance
type MongoDB struct {
	Session *mgo.Session
}

// Connect initiates a new MongoDB connection using the URL provided.
func Connect(url string) (MongoDB, error) {

	session, err := mgo.Dial(url)
	if err != nil {
		return MongoDB{}, err
	}
	return MongoDB{Session: session.Clone()}, nil
}
