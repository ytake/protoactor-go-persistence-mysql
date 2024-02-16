package persistencemysql

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid/v2"
	"github.com/ytake/protoactor-go-persistence-mysql/testdata"
)

func mysqlConfig() mysql.Config {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	return mysql.Config{
		DBName:    "sample",
		User:      "user",
		Passwd:    "passw@rd",
		Addr:      "localhost:3306",
		Net:       "tcp",
		ParseTime: true,
		Collation: "utf8mb4_bin",
		Loc:       jst,
	}
}

func TestProvider_PersistEvent(t *testing.T) {
	config := mysqlConfig()
	db, _ := sql.Open("mysql", config.FormatDSN())
	t.Cleanup(func() {
		db.Exec("TRUNCATE journals")
		db.Close()
	})
	provider, _ := New(3, NewTable(), db, nil)
	evt := &testdata.UserCreated{
		UserID:   ulid.Make().String(),
		UserName: "test",
		Email:    "",
	}
	provider.PersistEvent("user", 1, evt)
	var evv *testdata.UserCreated
	provider.GetEvents("user", 1, 4, func(e interface{}) {
		ev, ok := e.(*testdata.UserCreated)
		if !ok {
			t.Error("unexpected type")
		}
		evv = ev
	})
	if !reflect.DeepEqual(evt, evv) {
		t.Errorf("unexpected event %v", evv)
	}

	var evv2 *testdata.UserCreated
	provider.GetEvents("user", 1, 0, func(e interface{}) {
		ev, ok := e.(*testdata.UserCreated)
		if !ok {
			t.Error("unexpected type")
		}
		evv2 = ev
	})
	if !reflect.DeepEqual(evt, evv2) {
		t.Errorf("unexpected event %v", evv2)
	}
}

func TestProvider_PersistSnapshot(t *testing.T) {
	config := mysqlConfig()
	db, _ := sql.Open("mysql", config.FormatDSN())
	t.Cleanup(func() {
		db.Exec("TRUNCATE snapshots")
		db.Close()
	})
	provider, _ := New(3, NewTable(), db, nil)
	evt := &testdata.UserCreated{
		UserID:   ulid.Make().String(),
		UserName: "test",
		Email:    "",
	}
	provider.PersistSnapshot("user", 1, evt)
	snapshot, idx, ok := provider.GetSnapshot("user")
	if !ok {
		t.Error("snapshot not found")
	}
	if idx != 1 {
		t.Errorf("unexpected index %d", idx)
	}
	if !reflect.DeepEqual(snapshot, evt) {
		t.Errorf("unexpected snapshot %v", snapshot)
	}
}
