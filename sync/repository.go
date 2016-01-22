package sync

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

const connection_string = "root:root123@/syncmysport?charset=utf8,parseTime=true"

type DbSync struct {
}

func CreateSyncDbRepo() *DbSync {
	return &DbSync{}
}

func (db DbSync) storeSync(sync SyncTask) (int64, int64, error) {
	dbCon, _ := sql.Open("mysql", connection_string)
	defer dbCon.Close()

	stmtOut, err := dbCon.Prepare("INSERT INTO sync(rk_token, stv_token, last_succesfull_retrieve) VALUES(?,?,?)")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()
	res, err := stmtOut.Exec(sync.RunkeeperToken, sync.StravaToken, sync.LastSeenTimestamp)
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	return lastId, rowCnt, nil
}

func (db DbSync) RetrieveSyncTaskByToken(token string) (*SyncTask, error) {
	dbCon, _ := sql.Open("mysql", connection_string)
	stmtOut, err := dbCon.Prepare("SELECT * FROM sync WHERE rk_key = ? OR stv_key = ?")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(token, token)
	defer rows.Close()

	for rows.Next() {
		var uid int
		var rkToken string
		var stvToken string
		var lastSeen string
		err = rows.Scan(&uid, &rkToken, &stvToken, &lastSeen)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		log.Printf("lastseen: %s", lastSeen)
		//Mon Jan 2 15:04:05 -0700 MST 2006 <= Default format string
		timestamp, err := time.Parse("2006-01-02 15:04:05", lastSeen)

		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		task := CreateSyncTask(rkToken, stvToken, int(timestamp.Unix()))
		return task, nil
	}
	return nil, nil
}
