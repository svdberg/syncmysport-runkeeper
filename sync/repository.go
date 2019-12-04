package sync

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/go-sql-driver/mysql"
)

//should come from config (file) somewhere...
const default_connection_string = "root:root123@/syncmysport?charset=utf8,parseTime=true"

type DbSyncInt interface {
	UpdateSyncTask(sync SyncTask) (int, error)
	StoreSyncTask(sync SyncTask) (int64, int64, SyncTask, error)
	RetrieveAllSyncTasks() ([]SyncTask, error)
	FindSyncTaskByToken(token string) (*SyncTask, error)
	CreateTableIfNotExist() error
	CountActiveUsers() (int, error)
}

type DbSync struct {
	ConnectionString string
}

func CreateSyncDbRepo(dbString string) DbSyncInt {
	if dbString != "" {
		dbString = makeDbStringHerokuCompliant(dbString)
		appendedConnectionString := fmt.Sprintf("%s", dbString)
		return &DbSync{appendedConnectionString}
	} else {
		return &DbSync{default_connection_string}
	}
}

func (db DbSync) CreateTableIfNotExist() error {
	dbCon, err := sql.Open("mysql", db.ConnectionString)
	if err != nil {
		return err
	}
	defer dbCon.Close()

	_, err = dbCon.Exec(`
	CREATE TABLE IF NOT EXISTS sync (
    uid INT(10) NOT NULL AUTO_INCREMENT,
    rk_key VARCHAR(64) NULL DEFAULT NULL,
    rk_refresh_token VARCHAR(64) DEFAULT NULL,
    last_succesfull_retrieve DATETIME NULL DEFAULT NULL,
	environment varchar(36) NOT NULL DEFAULT "Prod",
	stv_key VARCHAR(64) NULL DEFAULT NULL,
    stv_refresh_token VARCHAR(64) DEFAULT NULL,
    PRIMARY KEY (uid)
  );`)
	if err != nil {
		return err
	}
	log.Printf("Create table sync\n")
	return nil
}

func (db DbSync) CountActiveUsers() (int, error) {
	dbCon, _ := sql.Open("mysql", db.ConnectionString)
	defer dbCon.Close()
	rows, err := dbCon.Query("SELECT COUNT(*) FROM sync WHERE rk_key != '' AND rk_key IS NOT NULL AND stv_key != '' AND stv_key IS NOT NULL")

	if err != nil {
		return 0, err
	}
	count := 0
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return count, err
		}
	}
	defer rows.Close()

	return count, nil
}

func (db DbSync) UpdateSyncTask(sync SyncTask) (int, error) {
	if sync.Uid == -1 {
		return 0, errors.New("SyncTask was never stored before, use StoreSyncTask")
	}
	dbCon, _ := sql.Open("mysql", db.ConnectionString)
	defer dbCon.Close()

	stmtOut, err := dbCon.Prepare("UPDATE sync SET rk_key=?, stv_key=?, stv_refresh_token=?, last_succesfull_retrieve=? WHERE uid = ?")
	if err != nil {
		return 0, errors.New("Error preparing UPDATE statement for Task")
	}
	defer stmtOut.Close()

	res, err := stmtOut.Exec(sync.RunkeeperToken, sync.StravaToken, sync.StravaRefreshToken, createStringOutOfUnixTime(sync.LastSeenTimestamp), sync.Uid)
	if err != nil {
		return 0, errors.New("Error executing the UPDATE statement for Task")
	}

	i, err := res.RowsAffected()
	if err != nil {
		return 0, errors.New("Error reading rows affected after UPDATE")
	}
	return int(i), nil
}

/*
* Returns 1) Created Id, 2) Rows changed/added, 3)synctask, 4) error
 */
func (db DbSync) StoreSyncTask(sync SyncTask) (int64, int64, SyncTask, error) {
	dbCon, _ := sql.Open("mysql", db.ConnectionString)
	defer dbCon.Close()

	stmtOut, err := dbCon.Prepare("INSERT INTO sync(rk_key, stv_key, stv_refresh_token, last_succesfull_retrieve, environment) VALUES(?,?,?,?)")
	if err != nil {
		log.Printf("err: %s", err)
		return 0, 0, sync, err
	}
	defer stmtOut.Close()
	res, err := stmtOut.Exec(sync.RunkeeperToken, sync.StravaToken, sync.StravaRefreshToken, createStringOutOfUnixTime(sync.LastSeenTimestamp), sync.Environment)
	if err != nil {
		log.Printf("err: %s", err)
		return 0, 0, sync, err
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Printf("err: %s", err)
		return 0, 0, sync, err
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Printf("err: %s", err)
		return 0, 0, sync, err
	}
	sync.Uid = lastId
	return lastId, rowCnt, sync, nil
}

func (db DbSync) RetrieveAllSyncTasks() ([]SyncTask, error) {
	log.Printf("Connecting to DB using conn string %s", db.ConnectionString)
	dbCon, _ := sql.Open("mysql", db.ConnectionString)
	stmtOut, err :=
		dbCon.Prepare("SELECT uid, rk_key, stv_key, last_succesfull_retrieve, environment, stv_refresh_token, rk_refresh_token FROM sync WHERE rk_key != '' AND stv_key != ''")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()
	defer dbCon.Close()

	rows, err := stmtOut.Query()
	defer rows.Close()

	result := make([]SyncTask, 0)
	for rows.Next() {
		var rkToken string
		var stvToken string
		var rkRefreshToken string
		var stvRefreshToken string
		var uid int64
		var lastSeenTime string
		var environment string

		rows.Scan(&uid, &rkToken, &stvToken, &lastSeenTime, &environment, &stvRefreshToken, &rkRefreshToken)
		unixTime, err := createUnixTimeOutOfString(lastSeenTime)
		if err != nil {
			log.Printf("Error while loading Sync tasks from database: %s", err)
			return result, err // proper error handling instead of panic in your app
		}

		sync := CreateSyncTask(rkToken, stvToken, rkRefreshToken, stvRefreshToken, unixTime, environment)
		sync.Uid = uid
		result = append(result, *sync)
	}
	return result, nil
}

func (db DbSync) FindSyncTaskByToken(token string) (*SyncTask, error) {
	dbCon, _ := sql.Open("mysql", db.ConnectionString)
	defer dbCon.Close()
	stmtOut, err := dbCon.Prepare("SELECT * FROM sync WHERE rk_key = ? OR stv_key = ? OR stv_refresh_token = ?")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(token, token, token)
	defer rows.Close()

	for rows.Next() {
		var uid int64
		var rkToken string
		var stvToken string
		var lastSeen string
		var environment string
		var stv_refresh_token string
		var rk_refresh_token string

		err = rows.Scan(&uid, &rkToken, &stvToken, &lastSeen, &environment, &stv_refresh_token, &rk_refresh_token)
		if err != nil {
			log.Printf("Error while getting results from db for token %s", token)
			return nil, err
		}
		unixTime, err := createUnixTimeOutOfString(lastSeen)
		if err != nil {
			log.Printf("Error while converting timestamp from db %s", lastSeen)
			return nil, err
		}
		task := CreateSyncTask(rkToken, stvToken, rk_refresh_token, stv_refresh_token, unixTime, environment)
		task.Uid = uid
		return task, nil
	}
	return nil, nil
}

func makeDbStringHerokuCompliant(dbString string) string {
	dbStringWithoutProtocol := strings.Replace(dbString, "mysql://", "", 1)
	parts := strings.Split(dbStringWithoutProtocol, "@")
	userAndPassword := strings.Split(parts[0], ":")

	addr := strings.Split(parts[1], "/")[0]
	dbName := strings.Split(strings.Split(parts[1], "/")[1], "?")[0]

	resultString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", userAndPassword[0], userAndPassword[1], addr, dbName)
	return resultString
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
