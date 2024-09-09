package db

import (
	"context"
	"testing"
)

type testEntity struct {
	IDField string `json:"-"`
	Field   string `json:"field"`
}

func (e *testEntity) SetID(s string) {
	e.IDField = s
}

func (e *testEntity) ID() string {
	return e.IDField
}

func (*testEntity) Type() string {
	return "test"
}

func TestNew(t *testing.T) {
	ctx := context.Background()

	db, err := New("test.data")
	if err != nil {
		t.Errorf("setup err: %v", err)
		t.FailNow()
	}

	t.Logf("%#v", db)

	e2 := &testEntity{
		IDField: "id2",
	}

	byt1, err := db.LoadRaw(ctx, "test", "id1")
	if err != nil {
		t.Errorf("load err: %v", err)
	}

	t.Log("bytes", string(byt1))

	byt2, err := db.LoadRaw(ctx, "test", "id2")
	if err != nil {
		t.Errorf("load err: %v", err)
	}

	t.Log("bytes", string(byt2))

	err = db.Load(ctx, e2)
	if err != nil {
		t.Errorf("load err: %v", err)
	}

	if e2.Field != "value 2" {
		t.Errorf("value err: %v", e2.Field)
	}

	e4 := &testEntity{
		IDField: "id4",
		Field:   "value 4",
	}

	err = db.Store(ctx, e4)
	if err != nil {
		t.Errorf("store err: %v", err)
	}

	e4 = &testEntity{
		IDField: "id4",
	}

	err = db.Load(ctx, e4)
	if err != nil {
		t.Errorf("load err: %v", err)
	}

	if e4.Field != "value 4" {
		t.Errorf("value err: %v", e4.Field)
	}

	t.Logf("%#v", db)

	t.Fail()
}
