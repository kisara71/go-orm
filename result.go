package go_orm

import "database/sql"

type ExecResult struct {
	res sql.Result
	err error
}

func (r *ExecResult) LastInsertID() (int64, error) {
	if r.err != nil {
		return 0, nil
	}
	return r.res.LastInsertId()
}
func (r *ExecResult) RowsAffected() (int64, error) {
	if r.err != nil {
		return 0, nil
	}
	return r.res.RowsAffected()
}
func (r *ExecResult) Err() error {
	return r.err
}
