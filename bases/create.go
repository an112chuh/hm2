package bases

import (
	"database/sql"
	"hm2/report"
	"net/http"
)

func CreateBases(r *http.Request, tx *sql.Tx, IDTeam int) (*sql.Tx, error) {
	var err error
	tx, err = CreateTrainingCenter(r, tx, IDTeam)
	if err != nil {
		return tx, err
	}
	tx, err = CreateScoutCenter(r, tx, IDTeam)
	if err != nil {
		return tx, err
	}
	tx, err = CreateMedCenter(r, tx, IDTeam)
	if err != nil {
		return tx, err
	}
	tx, err = CreatePhysCenter(r, tx, IDTeam)
	if err != nil {
		return tx, err
	}
	tx, err = CreateAcademy(r, tx, IDTeam)
	if err != nil {
		return tx, err
	}
	tx, err = CreatePsyCenter(r, tx, IDTeam)
	return tx, err
}

func CreateTrainingCenter(r *http.Request, tx *sql.Tx, IDTeam int) (*sql.Tx, error) {
	query := `insert into bases.training_center (team_id, lvl, in_build, days) VALUES ($1, 0, false, 0)`
	params := []interface{}{IDTeam}
	_, err := tx.Exec(query, params...)
	if err != nil {
		report.ErrorSQLServer(r, err, query, params...)
		return tx, err
	}
	return tx, err
}

func CreateScoutCenter(r *http.Request, tx *sql.Tx, IDTeam int) (*sql.Tx, error) {
	query := `insert into bases.scout_center (team_id, lvl, in_build, days) VALUES ($1, 0, false, 0)`
	params := []interface{}{IDTeam}
	_, err := tx.Exec(query, params...)
	if err != nil {
		report.ErrorSQLServer(r, err, query, params...)
		return tx, err
	}
	return tx, err
}

func CreateMedCenter(r *http.Request, tx *sql.Tx, IDTeam int) (*sql.Tx, error) {
	query := `insert into bases.med_center (team_id, lvl, in_build, days) VALUES ($1, 0, false, 0)`
	params := []interface{}{IDTeam}
	_, err := tx.Exec(query, params...)
	if err != nil {
		report.ErrorSQLServer(r, err, query, params...)
		return tx, err
	}
	return tx, err
}

func CreatePhysCenter(r *http.Request, tx *sql.Tx, IDTeam int) (*sql.Tx, error) {
	query := `insert into bases.phys_center (team_id, lvl, in_build, days) VALUES ($1, 0, false, 0)`
	params := []interface{}{IDTeam}
	_, err := tx.Exec(query, params...)
	if err != nil {
		report.ErrorSQLServer(r, err, query, params...)
		return tx, err
	}
	return tx, err
}

func CreateAcademy(r *http.Request, tx *sql.Tx, IDTeam int) (*sql.Tx, error) {
	query := `insert into bases.academy (team_id, lvl, in_build, days) VALUES ($1, 0, false, 0)`
	params := []interface{}{IDTeam}
	_, err := tx.Exec(query, params...)
	if err != nil {
		report.ErrorSQLServer(r, err, query, params...)
		return tx, err
	}
	return tx, err
}

func CreatePsyCenter(r *http.Request, tx *sql.Tx, IDTeam int) (*sql.Tx, error) {
	query := `insert into bases.psy_center (team_id, lvl, in_build, days) VALUES ($1, 0, false, 0)`
	params := []interface{}{IDTeam}
	_, err := tx.Exec(query, params...)
	if err != nil {
		report.ErrorSQLServer(r, err, query, params...)
		return tx, err
	}
	return tx, err
}
