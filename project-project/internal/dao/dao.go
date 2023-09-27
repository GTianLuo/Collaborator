package dao

import (
	"project-project/internal/database"
	"project-project/internal/database/gorms"
)

type TransactionUser struct {
	conn database.Dbconn
}

func (t TransactionUser) Action(f func(conn database.Dbconn) error) error {
	t.conn.Begin()
	err := f(t.conn)
	if err != nil {
		t.conn.Rollback()
		return err
	}
	t.conn.Commit()
	return nil
}

func NewTransactionUser() *TransactionUser {
	return &TransactionUser{
		conn: gorms.NewTran(),
	}
}
