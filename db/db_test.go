package db

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEntity struct {
	IDField ID     `json:"-"`
	Field   string `json:"field"`
}

func (e *testEntity) SetID(s ID) {
	e.IDField = s
}

func (e *testEntity) ID() ID {
	return e.IDField
}

func (*testEntity) Type() ID {
	return Mkid("test")
}

func dumpData(filename string, data string) error {
	return os.WriteFile(filename, []byte(data), 0600)
}

func TestNew(t *testing.T) {
	ctx := context.Background()

	startData := `test id1 {"field":"value 1 ................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................."}
test id2 {"field":"value 2"}
test id3 {"field":"value 3"}
test id4 {"field":"value 4"}
test id5 {"field":"value 4"}
test id4 {"field":"value 4"}
`

	err := dumpData("test.data", startData)
	require.NoError(t, err, "dump")

	db, err := New("test.data")
	require.NoError(t, err, "setup")

	t.Logf("%#v", db)

	e2 := &testEntity{
		IDField: Mkid("id2"),
	}

	// load raw 1

	_, byt1, err := db.loadRaw(Mkid("test"), Mkid("id1"))
	require.NoError(t, err, "load raw 1")

	t.Log("bytes", string(byt1))

	// load raw 2

	_, byt2, err := db.loadRaw(Mkid("test"), Mkid("id2"))
	require.NoError(t, err, "load raw 2")

	t.Log("bytes", string(byt2))

	// load 2

	err = db.Load(ctx, e2)
	require.NoError(t, err, "load 2")

	assert.Equal(t, "value 2", e2.Field)

	// store 4

	e4 := &testEntity{
		IDField: Mkid("id4"),
		Field:   "value 4",
	}

	err = db.Store(ctx, e4)
	require.NoError(t, err, "store 4")

	// load 4

	e4 = &testEntity{
		IDField: Mkid("id4"),
	}

	err = db.Load(ctx, e4)
	require.NoError(t, err, "load 4")

	assert.Equal(t, "value 4", e4.Field)

	t.Logf("%#v", db)
}

func TestIndex(t *testing.T) {
	ctx := context.Background()

	startData := `test id1 {"field":"value 1"}
test id2 {"field":"value 2"}
test id3 {"field":"value 3"}
test id4 {"field":"value 4"}
test id5 {"field":"value 4"}
test id4 {"field":"value 4"}
`

	err := dumpData("test.data", startData)
	require.NoError(t, err, "dump")

	db, err := New("test.data")
	require.NoError(t, err, "setup")

	// index by value

	err = db.Index(ctx, &testEntity{}, Mkid("idx1"), func(x Entity) ID {
		return Mkid(x.(*testEntity).Field)
	})
	require.NoError(t, err, "index")

	// query by value

	var list []testEntity

	err = db.Query(ctx, Mkid("idx1"), Mkid("value 4"), &list)
	require.NoError(t, err, "query")

	assert.Len(t, list, 2, "result list: %v", list)

	// add a new entity, with same value

	e100 := testEntity{
		IDField: Mkid("id100"),
		Field:   "value 4",
	}

	err = db.Store(ctx, &e100)
	require.NoError(t, err, "store")

	// query by value again

	err = db.Query(ctx, Mkid("idx1"), Mkid("value 4"), &list)
	require.NoError(t, err, "query")

	assert.Len(t, list, 3, "result list: %v", list)
}
