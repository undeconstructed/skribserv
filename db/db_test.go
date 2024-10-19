package db

import (
	"context"
	"fmt"
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

	goodData1 := `test id1 {"field":"value 1"}
test id2 {"field":"value 2"}
test id3 {"field":"value 3"}
test id4 {"field":"value 4.1"}
test id5 {"field":"value 5"}
test id4 {"field":"value 4.2"}
`
	goodData2 := `test id1 {"field":"value æŝðđŝ¶ŧĥŝ¶ĥĝ"}
test id2 {"field":"value 2"}
ŝ¶ŧĥ e¶æe€„ {"field":"value e¶ĝŝe€³ĝ"}
`
	badData1 := `test id1 {"field":"value 1"}
test id
test id3 asdasd
`
	badData2 := `test id1 {"field":"value 1"}

	test id2 {}
`
	badData3 := `test id1 {"field":"value 1"}
	test id2 {}`
	noData := ``

	versions := []struct {
		n string
		f func(string) (*DB, error)
	}{
		{"own", NewWithOwnCode},
		{"std", NewWithStdLib},
	}

	type test struct {
		name     string
		data     string
		err      string
		doReads  bool
		doWrites bool
	}

	tests := []test{
		{name: "good1", data: goodData1, doReads: true, doWrites: true},
		{name: "good2", data: goodData2, doReads: true, doWrites: true},
		{name: "blank", data: noData, doWrites: true},
		{name: "bad1", data: badData1, err: "line @ 2"},
		{name: "bad2", data: badData2, err: "line @ 2"},
		{name: "bad3", data: badData3, err: "line @ 2"},
	}

	makeTest := func(f func(string) (*DB, error), test test) func(*testing.T) {
		return func(t *testing.T) {
			err := dumpData("test.data", test.data)
			require.NoError(t, err, "dump")

			db, err := f("test.data")

			if test.err != "" {
				require.ErrorContains(t, err, test.err, "setup")
				return
			} else {
				require.NoError(t, err, "setup")
			}

			t.Logf("%#v", db)

			if test.doReads {
				// load raw 1

				_, byt1, err := db.loadRaw(Mkid("test"), Mkid("id1"))
				require.NoError(t, err, "load raw 1")

				t.Log("bytes", string(byt1))

				// load raw 2

				_, byt2, err := db.loadRaw(Mkid("test"), Mkid("id2"))
				require.NoError(t, err, "load raw 2")

				t.Log("bytes", string(byt2))

				// load 2

				e2 := &testEntity{
					IDField: Mkid("id2"),
				}

				err = db.Load(ctx, e2)
				require.NoError(t, err, "load 2")

				assert.Equal(t, "value 2", e2.Field)
			}

			if test.doWrites {
				// store 4

				e4 := &testEntity{
					IDField: Mkid("id4"),
					Field:   "value 4.3",
				}

				err = db.Store(ctx, e4)
				require.NoError(t, err, "store 4")

				// load 4

				e4 = &testEntity{
					IDField: Mkid("id4"),
				}

				err = db.Load(ctx, e4)
				require.NoError(t, err, "load 4")

				assert.Equal(t, "value 4.3", e4.Field)
			}

			t.Logf("%#v", db)
		}
	}

	for _, test := range tests {
		for _, version := range versions {
			t.Run(fmt.Sprintf("%v/%s", version.n, test.name), makeTest(version.f, test))
		}
	}
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

	db, err := NewWithOwnCode("test.data")
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

func TestCompact(t *testing.T) {
	ctx := context.Background()

	startData := `test id1 {"field":"value 1.1"}
test id2 {"field":"value 2.1"}
test id1 {"field":"value 1.2"}
test id4 {"field":"value 4.1"}
test id5 {"field":"value 5.1"}
test id4 {"field":"value 4.2"}
`

	err := dumpData("test.data", startData)
	require.NoError(t, err, "dump")

	db, err := NewWithOwnCode("test.data")
	require.NoError(t, err, "setup")

	err = db.Compact(ctx)
	require.NoError(t, err, "compact")

	// load 4

	e4 := &testEntity{
		IDField: Mkid("id4"),
	}

	err = db.Load(ctx, e4)
	require.NoError(t, err, "load 4")

	assert.Equal(t, "value 4.2", e4.Field)

	// store 2

	e2 := &testEntity{
		IDField: Mkid("id2"),
		Field:   "value 2.2",
	}

	err = db.Store(ctx, e2)
	require.NoError(t, err, "store 2")

	// load 2

	e2 = &testEntity{
		IDField: Mkid("id2"),
	}

	err = db.Load(ctx, e2)
	require.NoError(t, err, "load 2")

	assert.Equal(t, "value 2.2", e2.Field)
}
