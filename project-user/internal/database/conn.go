package database

type Dbconn interface {
	Rollback()
	Commit()
	Begin()
}
