package go_orm

import "database/sql"

type Result struct {
	res sql.Result
	err error
}

func (r *Result) LastInsertID() (int64, error) {
	if r.err != nil {
		return 0, nil
	}
	return r.res.LastInsertId()
}
func (r *Result) RowsAffected() (int64, error) {
	if r.err != nil {
		return 0, nil
	}
	return r.res.RowsAffected()
}
func (r *Result) Err() error {
	return r.err
}
