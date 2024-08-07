package databases

import (
	"database/sql"
	"github.com/Re-Wi/GoKitReWi/logger"
	_ "github.com/lib/pq"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite" // Sqlite driver based on GGO
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"time"
)

var Db *gorm.DB

type DbGorm struct {
	Type         string
	Dsn          string
	MaxIdleConns int
	MaxOpenConns int
}

func (dg *DbGorm) GormInit() *gorm.DB {
	switch dg.Type {
	case "mysql":
		Db = dg.GormMysql()
	case "postgresql":
		Db = dg.GormPostgresql()
	case "sqlite3":
		Db = dg.GormSqlite3()
	}
	return Db
}

func (dg *DbGorm) GormMysql() *gorm.DB {

	logger.Log.Infof("连接mysql [%s]", dg.Dsn)
	mysqlConfig := mysql.Config{
		DSN:                       dg.Dsn, // DSN data source name
		DefaultStringSize:         191,    // string 类型字段的默认长度
		DisableDatetimePrecision:  true,   // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,   // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,   // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,  // 根据版本自动配置
	}
	ormConfig := &gorm.Config{Logger: gormlog.Default.LogMode(gormlog.Silent)}
	gormDb, err := gorm.Open(mysql.New(mysqlConfig), ormConfig)
	if err != nil {
		logger.Log.Panicf("连接mysql失败! [%s]", err.Error())
		return nil
	}
	sqlDB, _ := gormDb.DB()
	sqlDB.SetMaxIdleConns(dg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dg.MaxOpenConns)
	return gormDb
}

func (dg *DbGorm) GormPostgresql() *gorm.DB {

	db, err := sql.Open("postgres", dg.Dsn)
	if err != nil {
		logger.Log.Panicf("连接postgresql失败! [%s]", err.Error())
	}
	gormDb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		logger.Log.Panicf("连接postgresql失败! [%s]", err.Error())
	}
	sqlDB, err := gormDb.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(dg.MaxIdleConns)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(dg.MaxOpenConns)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	return gormDb
}

func (dg *DbGorm) GormSqlite3() *gorm.DB {
	gormDb, err := gorm.Open(sqlite.Open(dg.Dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Panicf("连接Sqlite3失败! [%s]", err.Error())
	}
	sqlDB, err := gormDb.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(dg.MaxIdleConns)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(dg.MaxOpenConns)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	return gormDb
}
