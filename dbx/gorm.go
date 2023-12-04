package dbx

import (
	"context"
	"database/sql"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"k8s.io/klog/v2"
)

type DBConfig struct {
	DBHost      string `env:"DB_HOST,required" envDefault:"localhost"`
	DBPort      int    `env:"DB_PORT,required" envDefault:"5432"`
	DBName      string `env:"DB_NAME,required" envDefault:"postgres"`
	DBUser      string `env:"DB_USER,required" envDefault:"postgres"`
	DBPass      string `env:"DB_PASS,required"`
	DBTelemetry bool   `env:"DB_TELEMETRY"`
	DBDebug     bool   `env:"DB_DEBUG"`
	DBInit      bool   `env:"DB_INIT"`
}

type GormHelper struct {
	db    *gorm.DB
	sqlDB *sql.DB
	debug bool
}

func (x *GormHelper) Open(cfg *DBConfig, provider DBProvider) (err error) {
	if cfg.DBDebug {
		klog.Info("GormHelper debug on")
		x.debug = true
	}
	if x.db != nil {
		klog.Warningf("GormHelper opened, will close old connection")
	}
	klog.V(1).Infof("connection to db: %s@%s:%d/%s",
		cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	x.db, err = gorm.Open(provider(cfg), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		klog.Errorf("failed to connect db: %+v", cfg)
		return
	}
	if cfg.DBTelemetry {
		klog.Info("GormHelper telemetry on")
		if err := x.db.Use(otelgorm.NewPlugin()); err != nil {
			klog.Errorf("failed to setup telemetry: %s", err.Error())
		}
	}
	x.sqlDB, err = x.db.DB()
	return
}

func (x *GormHelper) Close() {
	if x.sqlDB == nil {
		return
	}
	if err := x.sqlDB.Close(); err != nil {
		klog.Errorf("failed to close db: %v", err)
	}
}

func (x *GormHelper) DB() *gorm.DB {
	if x.debug {
		return x.db.Debug()
	}
	return x.db
}

func (x *GormHelper) DBWithCtx(ctx context.Context) *gorm.DB {
	db := x.db.WithContext(ctx)
	if x.debug {
		return db.Debug()
	}
	return db
}
