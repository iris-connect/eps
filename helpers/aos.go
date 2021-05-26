// IRIS Endpoint-Server (EPS)
// Copyright (C) 2021-2021 The IRIS Endpoint-Server Authors (see AUTHORS.md)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package helpers

// Adapted from https://github.com/kiprotect/kodex/blob/master/parameters/file.go

/*
This implements an append-only file-based storage mechanism. It splits data
entries into chunks that are below the PIPE_BUF size limit, and thus will be
atomic on *most* POSIX-compliant file systems. This makes it possible to use
multiple instances of the storage mechanism in parallel, as the store will be
able to  reassemble individual data chunks even if their order has been mangled
by parallel writes from different programs.

You should not use this to write data to network file systems like NFS, or other
file systems that do not implement proper atomic APPEND operations.

This storage mechanism is well suited for simple, append-only operations. You
can implement a compaction by e.g. regularly switching file names and rewriting
an in-memory copy to the new file. Again, parallel writing will not compromise
the file's integrity.
*/

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/iris-connect/eps"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const BUFFER_SIZE = 255
const CHUNK_ID_LENGTH = 16
const CHUNK_HEADER_SIZE = 8 + CHUNK_ID_LENGTH
const CHUNK_VERSION = 1
const ENTRY_VERSION = 1

const (
	NullType = 0
)

type DataEntry struct {
	Type uint8
	ID   []byte
	Data []byte
}

type DataChunk struct {
	// The number of chunks for this hash
	Chunks uint16
	// The index of this chunk
	Index uint16
	// The actual ID of the entry this chunk belongs to
	ID []byte
	// The actual data in this chunk
	Data []byte
}

func (d *DataChunk) Read(reader io.Reader) error {

	d.ID = nil
	d.Data = nil
	d.Chunks = 0
	d.Index = 0

	bs2 := make([]byte, 2)
	bs1 := make([]byte, 1)

	// version
	if _, err := io.ReadFull(reader, bs1); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	version := bs1[0]

	if version != 1 {
		return fmt.Errorf("unknown version: %d", version)
	}

	// chunks
	if _, err := io.ReadFull(reader, bs2); err != nil {
		return err
	}

	d.Chunks = binary.LittleEndian.Uint16(bs2)

	// index
	if _, err := io.ReadFull(reader, bs2); err != nil {
		return err
	}

	d.Index = binary.LittleEndian.Uint16(bs2)

	// id length
	if _, err := io.ReadFull(reader, bs2); err != nil {
		return err
	}

	idLength := binary.LittleEndian.Uint16(bs2)

	d.ID = make([]byte, idLength)

	// id
	if _, err := io.ReadFull(reader, d.ID); err != nil {
		return err
	}

	if _, err := io.ReadFull(reader, bs2); err != nil {
		return err
	}

	dataLength := binary.LittleEndian.Uint16(bs2)

	d.Data = make([]byte, dataLength)

	// data
	if _, err := io.ReadFull(reader, d.Data); err != nil {
		return err
	}

	return nil
}

// Writes a data chunkn to the given writer
func (d *DataChunk) Write(writer io.Writer) error {
	var buffer bytes.Buffer
	// we write the version
	if size, err := buffer.Write([]byte{CHUNK_VERSION}); err != nil {
		return err
	} else if size != 1 {
		return fmt.Errorf("could not write")
	}
	// we write the number of chunks
	bs2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs2, d.Chunks)
	if size, err := buffer.Write(bs2); err != nil {
		return err
	} else if size != 2 {
		return fmt.Errorf("could not write")
	}
	// we write the chunk index
	binary.LittleEndian.PutUint16(bs2, d.Index)
	if size, err := buffer.Write(bs2); err != nil {
		return err
	} else if size != 2 {
		return fmt.Errorf("could not write")
	}
	// we write the ID length
	if len(d.ID) >= 1<<16 {
		return fmt.Errorf("ID is too long")
	}
	binary.LittleEndian.PutUint16(bs2, uint16(len(d.ID)))
	if size, err := buffer.Write(bs2); err != nil {
		return err
	} else if size != 2 {
		return fmt.Errorf("could not write")
	}

	if size, err := buffer.Write(d.ID); err != nil {
		return err
	} else if size != len(d.ID) {
		return fmt.Errorf("could not write")
	}

	// we write the data length
	binary.LittleEndian.PutUint16(bs2, uint16(len(d.Data)))
	if size, err := buffer.Write(bs2); err != nil {
		return err
	} else if size != 2 {
		return fmt.Errorf("could not write")
	}

	if size, err := buffer.Write(d.Data); err != nil {
		return err
	} else if size != len(d.Data) {
		return fmt.Errorf("could not write")
	}
	bytes := buffer.Bytes()
	if size, err := writer.Write(bytes); err != nil {
		return err
	} else if size != len(bytes) {
		return fmt.Errorf("could not write buffer")
	}
	return nil
}

func (e *DataEntry) ToBytes() []byte {
	bytes := make([]byte, len(e.Data)+len(e.ID)+4)
	bytes[0] = ENTRY_VERSION
	bytes[1] = e.Type
	binary.LittleEndian.PutUint16(bytes[2:4], uint16(len(e.ID)))
	copy(bytes[4:4+len(e.ID)], e.ID)
	copy(bytes[4+len(e.ID):len(bytes)], e.Data)
	return bytes
}

func (e *DataEntry) FromBytes(data []byte) error {
	if data[0] != ENTRY_VERSION {
		return fmt.Errorf("invalid version")
	}
	e.Type = data[1]
	idLength := binary.LittleEndian.Uint16(data[2:4])
	if int(idLength) > len(data)+4 {
		return fmt.Errorf("ID out of bounds")
	}
	e.ID = data[4 : 4+idLength]
	e.Data = data[4+idLength : len(data)]

	return nil

}

func MakeDataChunk(id []byte, chunks, index uint16, data []byte) *DataChunk {
	return &DataChunk{
		Chunks: chunks,
		Index:  index,
		Data:   data,
		ID:     id,
	}
}

// Splits a data entry into multiple data chunks.
func (e *DataEntry) Split() ([]*DataChunk, error) {
	bytes := e.ToBytes()
	effectiveBufferSize := BUFFER_SIZE - CHUNK_HEADER_SIZE
	chunks := len(bytes) / effectiveBufferSize
	if len(bytes)%effectiveBufferSize != 0 {
		chunks += 1
	}
	if chunks >= 1<<16 {
		return nil, fmt.Errorf("data is too large")
	}
	dataChunks := make([]*DataChunk, chunks)
	id, err := RandomID(CHUNK_ID_LENGTH)
	if err != nil {
		return nil, err
	}
	for i := 0; i < chunks; i++ {
		end := (i + 1) * effectiveBufferSize
		if end > len(bytes) {
			end = len(bytes)
		}
		dataChunks[i] = MakeDataChunk(
			id, uint16(chunks), uint16(i), bytes[i*effectiveBufferSize:end],
		)
	}
	eps.Log.Debugf("Split into %d entries", len(dataChunks))
	return dataChunks, nil
}

func (e *DataEntry) Reassemble(chunks []*DataChunk) error {
	e.Type = NullType
	e.Data = nil
	e.ID = nil
	data := make([]byte, 0, 100)
	for i, chunk := range chunks {
		if chunk.Index != uint16(i) {
			return nil
		}
		data = append(data, chunk.Data...)
	}

	if len(data) < 1 {
		return fmt.Errorf("invalid entry")
	}

	return e.FromBytes(data)
}

type ByPosition struct {
	Entries   []*DataEntry
	Positions map[string]int
}

func (b ByPosition) Len() int {
	return len(b.Entries)
}

func (b ByPosition) Swap(i, j int) {
	b.Entries[i], b.Entries[j] = b.Entries[j], b.Entries[i]
}

func (b ByPosition) Less(i, j int) bool {
	posI := b.Positions[string(b.Entries[i].ID)]
	posJ := b.Positions[string(b.Entries[j].ID)]
	return posI < posJ
}

// Reassembles data entries from a list of data chunks. Returns any remaining
// data chunks (which might be used later).
func reassemble(chunks []*DataChunk) ([]*DataEntry, []*DataChunk, error) {
	/*
		We can assume that chunks occur in the right order, but they still can
		be interleaved within each other.
	*/
	chunkPositions := make(map[string]int)
	entryPositions := make(map[string]int)
	dataEntries := make([]*DataEntry, 0, 10)
	remainingChunks := make([]*DataChunk, 0, 10)
	chunksByID := make(map[string][]*DataChunk)
	for i, chunk := range chunks {
		id := string(chunk.ID)
		if _, ok := chunkPositions[id]; !ok {
			chunkPositions[id] = i
		}
		existingChunks, ok := chunksByID[id]
		if ok {
			existingChunks = append(existingChunks, chunk)
		} else {
			existingChunks = make([]*DataChunk, 1, 10)
			existingChunks[0] = chunk
		}
		chunksByID[id] = existingChunks
	}
	for id, idChunks := range chunksByID {
		dataEntry := &DataEntry{}
		if err := dataEntry.Reassemble(idChunks); err != nil {
			return nil, nil, err
		}
		entryPositions[string(dataEntry.ID)] = chunkPositions[id]
		if dataEntry.ID == nil {
			remainingChunks = append(remainingChunks, idChunks...)
			continue
		}
		dataEntries = append(dataEntries, dataEntry)
	}
	sortedByPosition := &ByPosition{
		Entries:   dataEntries,
		Positions: entryPositions,
	}
	sort.Sort(sortedByPosition)
	eps.Log.Debugf("Found %d entries, %d remaining chunks", len(dataEntries), len(remainingChunks))
	return sortedByPosition.Entries, remainingChunks, nil
}

type DataStore interface {
	// Write data to the store
	Write(*DataEntry) error
	// Read data from the store
	Read() ([]*DataEntry, error)
	Init() error
}

// A file-based data store
type FileDataStore struct {
	filename string
	mutex    sync.Mutex
	wfile    *os.File
	rfile    *os.File
	chunks   []*DataChunk
}

func MakeFileDataStore(filename string) *FileDataStore {
	return &FileDataStore{
		filename: filename,
		mutex:    sync.Mutex{},
		chunks:   make([]*DataChunk, 0, 10),
	}
}

func (f *FileDataStore) Init() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// we try to create the directory if it doesn't exist
	dir := filepath.Dir(f.filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	wfile, err := os.OpenFile(f.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	f.wfile = wfile
	rfile, err := os.OpenFile(f.filename, os.O_CREATE|os.O_RDONLY, 0700)
	if err != nil {
		return err
	}
	f.rfile = rfile
	return nil
}

func (f *FileDataStore) readChunks() ([]*DataChunk, error) {
	chunks := make([]*DataChunk, 0, 10)
	for {
		chunk := &DataChunk{}
		position, err := f.rfile.Seek(0, 1)
		if err != nil {
			return nil, err
		}
		if err := chunk.Read(f.rfile); err != nil {
			if _, seekErr := f.rfile.Seek(position, 0); seekErr != nil {
				eps.Log.Errorf("Warning, two errors occured.")
				eps.Log.Error(seekErr)
			}
			return nil, err
		}
		if chunk.ID == nil {
			break
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

func (f *FileDataStore) Read() ([]*DataEntry, error) {

	f.mutex.Lock()
	defer f.mutex.Unlock()

	chunks, err := f.readChunks()
	if err != nil {
		return nil, err
	}
	if f.chunks != nil {
		chunks = append(f.chunks, chunks...)
	}
	dataEntries, remainingChunks, err := reassemble(chunks)
	if err != nil {
		return nil, err
	}
	f.chunks = remainingChunks
	return dataEntries, nil
}

func (f *FileDataStore) Write(entry *DataEntry) error {
	if chunks, err := entry.Split(); err != nil {
		return err
	} else {
		for _, chunk := range chunks {
			if err := chunk.Write(f.wfile); err != nil {
				return err
			}
		}
	}
	// we make sure the changes were all written to disk
	return f.wfile.Sync()
}
