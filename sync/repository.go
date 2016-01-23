package sync

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

//should come from config (file) somewhere...
const connection_string = "root:root123@/syncmysport?charset=utf8,parseTime=true"

type DbSync struct {
}

func CreateSyncDbRepo() *DbSync {
	return &DbSync{}
}

func (db DbSync) UpdateSyncTask(sync SyncTask) (int, error) {
	if sync.uid == -1 {
		return 0, errors.New("SyncTask was never stored before, use StoreSyncTask")
	}
	dbCon, _ := sql.Open("mysql", connection_string)
	defer dbCon.Close()

	stmtOut, err := dbCon.Prepare("UPDATE sync SET rk_key=?, stv_key=?, last_succesfull_retrieve=? WHERE uid = ?")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()
	res, err := stmtOut.Exec(sync.RunkeeperToken, sync.StravaToken, createStringOutOfUnixTime(sync.LastSeenTimestamp), sync.uid)
	if err != nil {
		log.Fatal(err)
	}

	i, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	return int(i), nil
}

func (db DbSync) StoreSyncTask(sync SyncTask) (int64, int64, error) {
	dbCon, _ := sql.Open("mysql", connection_string)
	defer dbCon.Close()

	stmtOut, err := dbCon.Prepare("INSERT INTO sync(rk_key, stv_key, last_succesfull_retrieve) VALUES(?,?,?)")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()
	res, err := stmtOut.Exec(sync.RunkeeperToken, sync.StravaToken, createStringOutOfUnixTime(sync.LastSeenTimestamp))
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

func (db DbSync) RetrieveAllSyncTasks() ([]SyncTask, error) {
	dbCon, _ := sql.Open("mysql", connection_string)
	stmtOut, err := dbCon.Prepare("SELECT * FROM sync")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query()
	defer rows.Close()

	result := make([]SyncTask, 0)
	for rows.Next() {
		var rkToken string
		var stvToken string
		var uid int64
		var lastSeenTime string

		rows.Scan(&uid, &rkToken, &stvToken, &lastSeenTime)
		unixTime, err := createUnixTimeOutOfString(lastSeenTime)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		sync := CreateSyncTask(rkToken, stvToken, unixTime)
		sync.uid = uid
		result = append(result, *sync)
	}
	return result, nil
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
		var uid int64
		var rkToken string
		var stvToken string
		var lastSeen string
		err = rows.Scan(&uid, &rkToken, &stvToken, &lastSeen)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		unixTime, err := createUnixTimeOutOfString(lastSeen)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		task := CreateSyncTask(rkToken, stvToken, unixTime)
		task.uid = uid
		return task, nil
	}
	return nil, nil
}

func createUnixTimeOutOfString(lastSeen string) (int, error) {
	timestamp, err := time.Parse("2006-01-02 15:04:05", lastSeen)

	if err != nil {
		return 0, err
	}
	return int(timestamp.Unix()), nil
}

func createStringOutOfUnixTime(t int) string {
	return time.Unix(int64(t), 0).Format("2006-01-02 15:04:05")
}
