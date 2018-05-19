package db

import (
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	"github.com/xormplus/xorm"
)

var defaultEngine *xorm.Engine

func OpenDB(debug bool, dialect, url string) (*xorm.Engine, error) {
	db, err := xorm.NewEngine(dialect, url)
	if err != nil {
		return nil, err
	}

	if err := KeepAlive(db); err != nil {
		return nil, err
	}

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}

	db.DatabaseTZ = location

	// if debug {
	// 	db.LogMode(true)
	// }

	//db.ShowSQL(true)
	return db, err
}

func Open(debug bool, dialect, dburl, reportdburl string) error {
	db, err := OpenDB(debug, dialect, dburl)
	if err != nil {
		return err
	}
	defaultEngine = db

	db, err = OpenDB(debug, dialect, reportdburl)
	if err != nil {
		return err
	}
	reportEngine = db

	return nil
}

func Session() *xorm.Session {
	s := defaultEngine.NewSession()
	s.IsAutoClose = true
	return s
}

