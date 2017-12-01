// All material is licensed under the Apache License Version 2.0, January 2004
// http://www.apache.org/licenses/LICENSE-2.0

package db

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
)

var dbName = os.Getenv("DB_NAME")

const (
	Users_Students_Collection      = "users_students"
	Users_Students_Meta_Collection = "users_students_meta"
	Master_Data_Collection         = "master_data"
	Latest_Fetch_Collection        = "latest_fetch"
)

// ErrInvalidDBProvided is returned in the event that an uninitialized db is
// used to perform actions against.
var ErrInvalidDBProvided = errors.New("invalid DB provided")

// DB is a collection of support for different DB technologies. Currently
// only MongoDB has been implemented. We want to be able to access the raw
// database support for the given DB so an interface does not work. Each
// database is too different.
type DB struct {

	// MongoDB Support.
	database *mgo.Database
	session  *mgo.Session
}

// NewMGO returns a new DB value for use with MongoDB based on a registered
// master session.
func NewMGO(url string, timeout time.Duration) (*DB, error) {
	ses, err := newMGO(url, timeout)
	if err != nil {
		return nil, errors.Wrapf(err, "mongo.New: %s,%v", url, timeout)
	}

	db := DB{
		database: ses.DB(dbName),
		session:  ses,
	}

	masterDataIndex := mgo.Index{
		Name:       "master_data_index",
		Key:        []string{"url"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = db.database.C(Master_Data_Collection).EnsureIndex(masterDataIndex)
	if err != nil {
		errors.Wrap(err, "<<unable to ensure index on masterDataIndex>>")
	}

	usersStudentsGoogleIDIndex := mgo.Index{
		Name:       "users_students_gID_index",
		Key:        []string{"google_id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = db.database.C(Users_Students_Collection).EnsureIndex(usersStudentsGoogleIDIndex)
	if err != nil {
		errors.Wrap(err, "<<unable to ensure index on usersStudentsGoogleIDIndex>>")
	}

	usersStudentsEmailIndex := mgo.Index{
		Name:       "users_students_email_index",
		Key:        []string{"email"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = db.database.C(Users_Students_Collection).EnsureIndex(usersStudentsEmailIndex)
	if err != nil {
		errors.Wrap(err, "<<unable to ensure index on usersStudentsEmailIndex>>")
	}

	usersStudentsMobileIndex := mgo.Index{
		Name:       "users_students_mobile_index",
		Key:        []string{"mobile"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = db.database.C(Users_Students_Collection).EnsureIndex(usersStudentsMobileIndex)
	if err != nil {
		errors.Wrap(err, "<<unable to ensure index on usersStudentsMobileIndex>>")
	}

	usersStudentsMetaUserIDIndex := mgo.Index{
		Name:       "users_students_meta_userID_index",
		Key:        []string{"user_id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = db.database.C(Users_Students_Meta_Collection).EnsureIndex(usersStudentsMetaUserIDIndex)
	if err != nil {
		errors.Wrap(err, "<<unable to ensure index on usersStudentsMetaUserIDIndex>>")
	}

	usersStudentsMetaGoogleIDIndex := mgo.Index{
		Name:       "users_students_meta_gID_index",
		Key:        []string{"google_id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = db.database.C(Users_Students_Meta_Collection).EnsureIndex(usersStudentsMetaGoogleIDIndex)
	if err != nil {
		errors.Wrap(err, "<<unable to ensure index on usersStudentsMetaGoogleIDIndex>>")
	}

	db.session.EnsureSafe(&mgo.Safe{W: 1, FSync: true})

	return &db, nil
}

// MGOClose closes a DB value being used with MongoDB.
func (db *DB) MGOClose() {
	db.session.Close()
}

// MGOCopy returns a new DB value for use with MongoDB based the master session.
func (db *DB) MGOCopy() (*DB, error) {
	ses := db.session.Copy()

	// As per the mgo documentation, https://godoc.org/gopkg.in/mgo.v2#Session.DB
	// if no database name is specified, then use the default one, or the one that
	// the connection was dialed with.
	newDB := DB{
		database: ses.DB(dbName),
		session:  ses,
	}

	return &newDB, nil
}

// MGOExecute is used to execute MongoDB commands.
func (db *DB) MGOExecute(ctx context.Context, collName string, f func(*mgo.Collection) error) error {
	if db == nil || db.session == nil {
		return errors.Wrap(ErrInvalidDBProvided, "db == nil || db.session == nil")
	}

	return f(db.database.C(collName))
}

// MGOExecuteTimeout is used to execute MongoDB commands with a timeout.
func (db *DB) MGOExecuteTimeout(ctx context.Context, collName string, f func(*mgo.Collection) error, timeout time.Duration) error {
	if db == nil || db.session == nil {
		return errors.Wrap(ErrInvalidDBProvided, "db == nil || db.session == nil")
	}

	db.session.SetSocketTimeout(timeout)

	return f(db.database.C(collName))
}

// newMGO creates a new mongo connection. If no url is provided,
// it will default to localhost:27017.
func newMGO(url string, timeout time.Duration) (*mgo.Session, error) {

	// Set the default timeout for the session.
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	ses, err := mgo.DialWithTimeout(url, timeout)
	if err != nil {
		return nil, errors.Wrapf(err, "mgo.DialWithTimeout: %s,%v", url, timeout)
	}

	// index := Index{
	// 	Key: []string{"url"},
	// 	Unique: true,
	// }

	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).
	// http://godoc.org/labix.org/v2/mgo#Session.SetMode
	ses.SetMode(mgo.Monotonic, true)

	return ses, nil
}

// Query provides a string version of the value
func Query(value interface{}) string {
	json, err := json.Marshal(value)
	if err != nil {
		return ""
	}

	return string(json)
}
