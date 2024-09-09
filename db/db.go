package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

var ErrNotFound = errors.New("not found")
var ErrCorrupt = errors.New("corrupt")

type Entity interface {
	ID() string
	Type() string
	SetID(s string)
}

type entry struct {
	species, id   string
	start, length int64
}

type DB struct {
	entries map[string]map[string]*entry

	file     *os.File
	position int
	size     int
	wasted   int
}

func New(path string) (*DB, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0700)
	if err != nil {
		return nil, err
	}

	// final results
	data := map[string]map[string]*entry{}
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
			// save the last line
			forSpecies, ok := data[species]
			if !ok {
				forSpecies = map[string]*entry{}
				data[species] = forSpecies
			}

			if old, ok := forSpecies[id]; ok {
				// old value found
				wasted += dataLength
				size -= int(old.length)
			}

			forSpecies[id] = &entry{
				species: species,
				id:      id,
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
		entries:  data,
		size:     size,
		wasted:   wasted,
		position: position,
	}, nil
}

func (db *DB) StoreRaw(ctx context.Context, species, id string, data []byte) error {
	meta := []byte(species + " " + id + " ")
	metaLength := len(meta)

	_, err := db.file.Write(meta)
	if err != nil {
		return err
	}

	_, err = db.file.Write(data)
	if err != nil {
		return err
	}

	_, err = db.file.Write([]byte("\n"))
	if err != nil {
		return err
	}

	dataStart := db.position + metaLength
	dataLength := len(data)

	// TODO deduplicate code
	forSpecies, ok := db.entries[species]
	if !ok {
		forSpecies = map[string]*entry{}
		db.entries[species] = forSpecies
	}

	if old, ok := forSpecies[id]; ok {
		// old value found
		db.wasted += dataLength
		db.size -= int(old.length)
	}

	forSpecies[id] = &entry{
		species: species,
		id:      id,
		start:   int64(dataStart),
		length:  int64(dataLength),
	}

	// update position for next insert
	db.position += metaLength + dataLength

	return nil
}

func (db *DB) Store(ctx context.Context, e Entity) error {
	species, id := e.Type(), e.ID()

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	err = db.StoreRaw(ctx, species, id, data)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) LoadRaw(ctx context.Context, species, id string) ([]byte, error) {
	entry, ok := db.entries[species][id]
	if !ok {
		return nil, ErrNotFound
	}

	buffer := make([]byte, entry.length)

	n, _ := db.file.ReadAt(buffer, entry.start)
	if n != int(entry.length) {
		return nil, ErrCorrupt
	}

	return buffer, nil
}

func (db *DB) Load(ctx context.Context, e Entity) error {
	species, id := e.Type(), e.ID()

	buffer, err := db.LoadRaw(ctx, species, id)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buffer, e)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCorrupt, err)
	}

	return nil
}

func (db *DB) Query(ctx context.Context, field, value string, es []Entity) error {
	return nil
}
