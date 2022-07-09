package database

type dbinterface interface {
	//creates the database if it does not exist
	Create()
}
