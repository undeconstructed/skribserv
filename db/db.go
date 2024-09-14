package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"unicode/utf8"
	"unique"
)

var ErrNotFound = errors.New("not found")
var ErrCorrupt = errors.New("corrupt")

type ID unique.Handle[string]

func Mkid(in string) ID {
	return ID(unique.Make(in))
}

func (i ID) String() string {
	return unique.Handle[string](i).Value()
}

func (i *ID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*i = ID(unique.Make(s))

	return nil
}

func (i ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

type Entity interface {
	ID() ID
	Type() ID
	SetID(s ID)
}

type entry struct {
	species, id   ID
	start, length int64
	indexed       []*idList
}

// species -> id -> entry
type entries map[ID]map[ID]*entry

type IndexerFunc func(e Entity) ID

type indexer struct {
	name ID
	fn   IndexerFunc
}

// species -> []indexer
type indexers map[ID][]indexer

type idList struct {
	ids []ID
}

// species -> index -> value -> []id
type indexes map[ID]map[ID]map[ID]*idList

type DB struct {
	// entries is main data table
	entries entries

	// indexers is definitions of indexes
	indexers indexers

	// indexes is indexed data
	indexes indexes

	// file for primary data
	file *os.File
	// position in bytes of end of data file
	position int
	// size in bytes of useful data (not metadata)
	size int
	// wasted size in bytes because of duplicates (not metadata)
	wasted int
}

func New(path string) (*DB, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0700)
	if err != nil {
		return nil, err
	}

	// final results
	entries := entries{}
	size := 0
	wasted := 0

	// for tracking the file as we read it
	buffer := make([]byte, 512)
	eof := false  // whether last read found EOF
	position := 0 // absolute position in file

	// initial read to fill the buffer
	n, err := f.ReadAt(buffer, 0)
	if err != nil {
		if errors.Is(err, io.EOF) {
			eof = true
		} else {
			return nil, err
		}
	}

	// readable contains the part of the buffer that is real data
	readable := buffer[:n]

	// for each line, we need to find these values
	species, id := "", ""
	lineStart, idStart, dataStart, dataLength, dataEnd := -1, 0, 0, 0, 0

	// do this loop while there is any data remaining, either from last read or left in the buffer
	for len(readable) > 0 {
		// want to range through the runes, so cast to string
		s := string(readable)

		// how much of the buffer is consumed until a line end
		consumed := 0

		for n, ch := range s {
			runeLen := utf8.RuneLen(ch)

			if lineStart == -1 {
				// have found the start of a line
				lineStart = position
			}

			if species == "" && ch == ' ' {
				// if we didn't have a species, we now have
				species = s[0:n]
				// ID starts after this space
				idStart = n + runeLen
			} else if id == "" && ch == ' ' {
				// if we didn't have an ID, we now have
				id = s[idStart:n]
				// data starts after this space, but we want the absolute position in the file
				dataStart = position + runeLen
			} else if ch == '\n' {
				// now we are at the end of the line
				if species == "" || id == "" {
					return nil, fmt.Errorf("corrupt 1")
				}

				dataEnd = position
				consumed = n + runeLen
				position += runeLen

				// breaking out the loop moves the post-line phase
				break
			} else if dataStart != 0 {
				dataLength += runeLen
			}

			position += runeLen
		}

		if dataEnd == 0 {
			// have not finished the line
			if eof {
				// but there is no more data
				return nil, fmt.Errorf("corrupt 2")
			}

			// otherwise mark the data as consumed and continue to the next buffer
			readable = nil
		} else {
			speciesH, idH := Mkid(species), Mkid(id)

			// save the last line
			forSpecies, ok := entries[speciesH]
			if !ok {
				forSpecies = map[ID]*entry{}
				entries[speciesH] = forSpecies
			}

			if old, ok := forSpecies[idH]; ok {
				// old value found
				wasted += dataLength
				size -= int(old.length)
			}

			forSpecies[idH] = &entry{
				species: speciesH,
				id:      idH,
				start:   int64(dataStart),
				length:  int64(dataLength),
			}

			size += dataLength

			// now get ready for the next line

			// copy remaining data back to the beginning of the buffer
			moved := copy(buffer, readable[consumed:])
			// remake readable to point to good data
			readable = buffer[0:moved]

			// reset the variables
			species, id = "", ""
			lineStart, idStart, dataStart, dataLength, dataEnd = -1, 0, 0, 0, 0
		}

		if !eof {
			offset := len(readable)

			// refill the buffer, and go round again
			n, err = f.ReadAt(buffer[offset:], int64(position))
			if err != nil {
				if errors.Is(err, io.EOF) {
					eof = true
				} else {
					return nil, err
				}
			}

			readable = buffer[0 : offset+n]
		}
	}

	return &DB{
		file:     f,
		entries:  entries,
		indexers: indexers{},
		indexes:  indexes{},
		size:     size,
		wasted:   wasted,
		position: position,
	}, nil
}

func (db *DB) Index(ctx context.Context, e Entity, name ID, indexFn IndexerFunc) error {
	speciesH, indexH := e.Type(), name

	// save the indexer for future changes

	db.indexers[speciesH] = append(db.indexers[speciesH], indexer{
		name: indexH,
		fn:   indexFn,
	})

	// index the existing data

	index := map[ID]*idList{} // value -> []id

	for _, entry := range db.entries[speciesH] {
		idH := entry.id

		data, err := db.readData(entry)
		if err != nil {
			return err
		}

		err = db.parse(idH, data, e)
		if err != nil {
			return err
		}

		value := indexFn(e)

		ids := index[value]
		if ids == nil {
			ids = &idList{}
			index[value] = ids
		}

		entry.indexed = append(entry.indexed, ids)

		ids.ids = append(ids.ids, idH)
	}

	forSpecies, ok := db.indexes[speciesH]
	if !ok {
		forSpecies = map[ID]map[ID]*idList{}
		db.indexes[speciesH] = forSpecies
	}

	forSpecies[indexH] = index

	return nil
}

func (db *DB) storeRaw(species, idH ID, data []byte) (*entry, error) {
	meta := []byte(species.String() + " " + idH.String() + " ")
	metaLength := len(meta)

	_, err := db.file.Write(meta)
	if err != nil {
		return nil, err
	}

	_, err = db.file.Write(data)
	if err != nil {
		return nil, err
	}

	_, err = db.file.Write([]byte("\n"))
	if err != nil {
		return nil, err
	}

	dataStart := db.position + metaLength
	dataLength := len(data)

	// TODO deduplicate code
	forSpecies, ok := db.entries[species]
	if !ok {
		forSpecies = map[ID]*entry{}
		db.entries[species] = forSpecies
	}

	oldEntry, hasOld := forSpecies[idH]

	if hasOld {
		// old value found
		db.wasted += dataLength
		db.size -= int(oldEntry.length)
	}

	forSpecies[idH] = &entry{
		species: species,
		id:      idH,
		start:   int64(dataStart),
		length:  int64(dataLength),
	}

	// update position for next insert
	db.position += metaLength + dataLength

	return oldEntry, nil
}

func (db *DB) Store(ctx context.Context, e Entity) error {
	speciesH, idH := e.Type(), e.ID()

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	old, err := db.storeRaw(speciesH, idH, data)
	if err != nil {
		return err
	}

	if old != nil {
		// remove old value from all indexes

		for _, list := range old.indexed {
			list.ids = slices.DeleteFunc(list.ids, func(e ID) bool {
				return e == idH
			})
		}
	}

	// apply indexers to new value

	for _, idxr := range db.indexers[speciesH] {
		index := db.indexes[speciesH][idxr.name]

		value := idxr.fn(e)

		ids := index[value]
		if ids == nil {
			ids = &idList{}
			index[value] = ids
		}

		ids.ids = append(ids.ids, idH)
	}

	return nil
}

func (db *DB) readData(entry *entry) ([]byte, error) {
	buffer := make([]byte, entry.length)

	n, _ := db.file.ReadAt(buffer, entry.start)
	if n != int(entry.length) {
		return nil, ErrCorrupt
	}

	return buffer, nil
}

func (db *DB) loadRaw(species, id ID) (*entry, []byte, error) {
	entry, ok := db.entries[species][id]
	if !ok {
		return nil, nil, ErrNotFound
	}

	data, err := db.readData(entry)
	if err != nil {
		return entry, nil, err
	}

	return entry, data, nil
}

func (db *DB) parse(idH ID, data []byte, e Entity) error {
	err := json.Unmarshal(data, e)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCorrupt, err)
	}

	e.SetID(idH)

	return nil
}

func (db *DB) Load(ctx context.Context, e Entity) error {
	speciesH, idH := e.Type(), e.ID()

	_, data, err := db.loadRaw(speciesH, idH)
	if err != nil {
		return err
	}

	return db.parse(idH, data, e)
}

func Query[E Entity](ctx context.Context, db *DB, index, value ID, es []E) error {
	return db.Query(ctx, index, value, es)
}

// Query
// es must be *[]E where E is an entity type
func (db *DB) Query(ctx context.Context, index, value ID, es any) error {
	rv := reflect.ValueOf(es) // *[]x
	rt := rv.Type()

	if rt.Kind() != reflect.Ptr {
		panic("must be a pointer to slice")
	}

	rv = rv.Elem() // []x
	rt = rt.Elem()

	if rt.Kind() != reflect.Slice {
		panic("must be a slice or pointer to a slice")
	}

	// replace with a new empty list
	newSlice := reflect.New(rt).Elem()
	rv.Set(newSlice)

	entityType := rt.Elem() // x
	species := reflect.New(entityType).Interface().(Entity).Type()

	speciesH, indexH, valueH := species, index, value

	list, ok := db.indexes[speciesH][indexH][valueH]
	if !ok {
		// no results, no change
		return nil
	}

	for _, idH := range list.ids {
		eRv := reflect.New(entityType)
		e := eRv.Interface().(Entity)

		e.SetID(idH)

		err := db.Load(ctx, e)
		if err != nil {
			return err
		}

		rv.Set(reflect.Append(rv, eRv.Elem()))
	}

	return nil
}
