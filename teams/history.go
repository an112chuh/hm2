package teams

type HistoryRecord struct {
	ID        int
	ManagerID int
	TeamName  string
	TeamID    int
	G         int
	W         int
	WO        int
	WS        int
	LS        int
	LO        int
	L         int
	Trophies  int
	IsFinish  bool
}

/*func SetHistoryRecord(tx *sql.Tx, hr HistoryRecord, ctx context.Context) (tx *sql.Tx, err error){
	query := `INSERT into managers.history (manager_id, date_start, team_name, team_id, G, W, WO, WS, LS, LO, L, trophies) VALUES ($1, $2, $3, $4, 0, 0, 0, 0, 0, 0, 0, 0)`
	t := time.Now().Format("02-01-2006")
	params := []interface{}{hr.ManagerID, t, hr.TeamName, hr.TeamID}
	_, err := tx.ExecContext(ctx, query, params...)




}

func UpdateHistoryRecord(hr HistoryRecord) {

} */
