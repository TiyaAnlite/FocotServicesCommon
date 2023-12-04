package dbx

import (
	"context"
	"fmt"
	"github.com/duke-git/lancet/v2/random"
	otelTrace "go.opentelemetry.io/otel/trace"
	"os"
	"testing"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type simpleTable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func TestOpel(t *testing.T) {
	ctx := context.Background()
	opelDSN := os.Getenv("TRACE_DSN")
	if opelDSN == "" {
		t.Error("openDSN not set")
		t.FailNow()
	}
	uptrace.ConfigureOpentelemetry(
		uptrace.WithDSN(opelDSN),
		uptrace.WithServiceName("GormHelper"),
		uptrace.WithServiceVersion("test"),
		uptrace.WithDeploymentEnvironment("test"),
	)
	defer uptrace.Shutdown(ctx)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Errorf("failed to init db: %s", err.Error())
		t.FailNow()
	}
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		t.Errorf("failed to init otel plguin: %s", err.Error())
		t.FailNow()
	}
	tracer := otel.Tracer("GormHelper-TestOpel@test")
	mainCtx, mainTracer := tracer.Start(ctx, "TestOpel")
	defer mainTracer.End()
	var ch otelTrace.Span
	var chCtx context.Context
	chCtx, ch = tracer.Start(mainCtx, "db init")
	if err := db.WithContext(chCtx).AutoMigrate(&simpleTable{}); err != nil {
		errMsg := fmt.Errorf("failed to migrate db: %s", err.Error())
		t.Errorf(errMsg.Error())
		ch.RecordError(err)
		ch.End()
		t.FailNow()
	}
	ch.End()

	chCtx, ch = tracer.Start(mainCtx, "data write")
	if err := db.WithContext(chCtx).Create(&simpleTable{
		Key:   "testingKey",
		Value: random.RandString(32),
	}).Error; err != nil {
		errMsg := fmt.Errorf("failed to write data: %s", err.Error())
		t.Errorf(errMsg.Error())
		ch.RecordError(err)
		ch.End()
		t.FailNow()
	}
	ch.End()

	chCtx, ch = tracer.Start(mainCtx, "data read")
	var readData simpleTable
	if err := db.WithContext(chCtx).Where("key = ?", "testingKey").Take(&readData).Error; err != nil {
		errMsg := fmt.Errorf("failed to read data: %s", err.Error())
		t.Errorf(errMsg.Error())
		ch.RecordError(err)
		ch.End()
		t.FailNow()
	}
	t.Logf("testingKey value: %s", readData.Value)
	ch.End()
}
