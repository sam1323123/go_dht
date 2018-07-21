package chordmap

import (
	"crypto/sha1"
	//"errors"
	K "go_dht/constants"
	SSA "go_dht/shasumarith"
)

type ChordMapStruct struct {
	start, end [K.ShaSize]byte // end is exclusive
	table      map[string]string
}

type CMInterface interface {
	Put(key string, value string) error
	Get(key string) (string, error)
	Delete(key string) (string, error)
}

/********* Helper functions **********************/

/*
Hashing function for a string
*/
func StrToSha(s string) [K.ShaSize]byte {
	return sha1.Sum([]byte(s))
}

/* Checks if x is within [start, end) i.e exclusive end, inclusive start. Also handles case of wrap around where end < start.
If end == start, returns true i.e everything is in range.
*/
func InRangeHelp(x [K.ShaSize]byte, start [K.ShaSize]byte, end [K.ShaSize]byte) bool {
	if SSA.Cmp(start, end) == SSA.Equal {
		return true
	}
	if SSA.Cmp(end, SSA.FromInt(0)) == SSA.Equal { // prevent subtract from 0 which results in compare with maxval
		return SSA.InRange(x, start, SSA.MaxVal())
	} else if SSA.Cmp(start, end) == SSA.Greater { // wrap around case and end != 0
		return SSA.InRange(x, start, SSA.MaxVal()) || InRangeHelp(x, SSA.FromInt(0), end)
		// check [start, MaxVal], [0, end)
	}
	// start < end
	return SSA.InRange(x, start, SSA.Sub(end, SSA.FromInt(1)))
}

/********** Interface implementations ***********/

/*
Inserts (key, value) into the cms.table if the sha sum is >= cms.start and < cms.end. Else return an erro
If key already present, updates with the newer entry
*/
func (cms *ChordMapStruct) Put(key string, value string) error {
	// if >= cms.start and < cms.end
	if !InRangeHelp(StrToSha(key), cms.start, cms.end) {
		return NewCMRangeError()
	}
	cms.table[key] = value
	return nil
}

/*
Gets values from table if key is within range of start and end and returns (value, nil)
Else if no key in table, return ("", CMKeyError)
Else returns ("", CMRangeError)
*/
func (cms *ChordMapStruct) Get(key string) (string, error) {
	if !InRangeHelp(StrToSha(key), cms.start, cms.end) {
		return "", NewCMRangeError()
	} else if ret, present := cms.table[key]; !present { // no key in table
		return "", NewCMKeyError()
	} else { // valid key with value inserted
		return ret, nil
	}
}

/*
Deletes key from table and returns (value, error)
If key is present return (value, nil)
Else if key is not in range return ("", CMRangeError)
Else if key nit present return ("", CMKeyError)
*/
func (cms *ChordMapStruct) Delete(key string) (string, error) {
	if !InRangeHelp(StrToSha(key), cms.start, cms.end) {
		return "", NewCMRangeError()
	} else if ret, present := cms.table[key]; !present { // no key in table
		return "", NewCMKeyError()
	} else { // valid key with value inserted
		delete(cms.table, key)
		return ret, nil
	}
}

/*
Returns all the keys in the chord map
*/
func (cms *ChordMapStruct) GetKeys() []string {
	keys := make([]string, len(cms.table))
	i := 0
	for k := range cms.table {
		keys[i] = k
		i++
	}
	return keys
}

/*
Splits a ChorMap on key. Returned chordmap gets all keys < key. cms gets remainder i.e [key, cms.end)
If key not [start, end) raise Range error. Else modifies cms and returns the extracted left half.
Returns 2 Chordmaps to of the correct range and entries
*/
func (cms *ChordMapStruct) PartitionTable(key [K.ShaSize]byte) (*ChordMapStruct, error) {
	if !InRangeHelp(key, cms.start, cms.end) {
		return nil, NewCMRangeError()
	}
	ret := New(key, cms.end)
	cms.end = key
	for k, v := range cms.table {
		err := ret.Put(k, v)
		if err != nil { // If key belongs in left table, delete key from cms
			cms.Delete(k)
		}
	}
	return ret, nil
}

/* initializes a new chord map struct. Inclusive start and exclusive end. If start == end, means chordmap accepts everything
Handles wrap arounds for start and end
*/
func New(start [K.ShaSize]byte, end [K.ShaSize]byte) *ChordMapStruct {
	ret := new(ChordMapStruct)
	copy(ret.start[:], start[:]) // make copy of array
	copy(ret.end[:], end[:])
	ret.table = make(map[string]string)
	return ret
}
