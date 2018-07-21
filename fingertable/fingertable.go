package fingertable

import (
	"go_dht/constants"
	SSA "go_dht/shasumarith"
)

type HostStruct struct {
	Hostname, Port string
}

type FTStruct struct {
	n     [constants.ShaSize]byte          // the ending key for the finger table and this node
	table [constants.ShaNumBits]HostStruct // maps i -> (host, port). Each entry stores succ(n + 2^i)
}

type UpdateFn func([constants.ShaSize]byte) (string, string)

/************ Helper Functions ***************/

/* Checks if x is within [start, end) i.e exclusive end, inclusive start. Also handles case of wrap around where end < start.
If end == start, returns false.
*/
func InRangeHelp(x [constants.ShaSize]byte, start [constants.ShaSize]byte, end [constants.ShaSize]byte) bool {
	if SSA.Cmp(start, end) == SSA.Equal {
		return false
	}
	if SSA.Cmp(start, end) == SSA.Greater { // wrap around case
		return SSA.InRange(x, start, SSA.MaxVal()) || InRangeHelp(x, SSA.FromInt(0), end)
		// check [start, MaxVal], [0, end)
	}
	// start < end
	return SSA.InRange(x, start, SSA.Sub(end, SSA.FromInt(1)))
}

/************** End Helper *********************/

/*
Given a sha key, returns the host and port of the node responsible for it
*/
func (fts *FTStruct) Find(key [constants.ShaSize]byte) (string, string, error) {
	start := SSA.Add(fts.n, SSA.Pow2(0)) // n + 2^0
	var end [constants.ShaSize]byte
	for i := uint32(0); i < constants.ShaNumBits; i++ {
		if i == constants.ShaNumBits-1 { // wrap around
			end = SSA.Add(fts.n, SSA.Pow2(0)) // n + 2^0
		} else {
			end = SSA.Add(fts.n, SSA.Pow2(i+1)) // n + s^(i+1)
		}
		if InRangeHelp(key, start, end) {
			return fts.table[i].Hostname, fts.table[i].Port, nil
			// return start/more-left node as the predecessor is better candidate since search is done clockwise
		}
		start = end // update start
	}
	return "", "", NewFTFindError()
}

/*
Given a sha key, returns the corresponding index of the finger table
*/
func (fts *FTStruct) FindIndex(key [constants.ShaSize]byte) (uint32, error) {
	start := SSA.Add(fts.n, SSA.Pow2(0)) // n + 2^0
	var end [constants.ShaSize]byte
	for i := uint32(0); i < constants.ShaNumBits; i++ {
		if i == constants.ShaNumBits-1 { // wrap around
			end = SSA.Add(fts.n, SSA.Pow2(0)) // n + 2^0
		} else {
			end = SSA.Add(fts.n, SSA.Pow2(i+1)) // n + s^(i+1)
		}
		if InRangeHelp(key, start, end) {
			return i, nil
		}
		start = end // update start, less calculation
	}
	return 0, NewFTFindError()
}

// updatesthe entire fingertable using u_fn
func (fts *FTStruct) Update(u_fn UpdateFn) {
	var new_tab [constants.ShaNumBits]HostStruct // required in case u_fn uses functions that reads/writes to fts.table
	for i := range fts.table {
		name, port := u_fn(SSA.Add(fts.n, SSA.Pow2(uint32(i))))
		new_tab[i] = HostStruct{Hostname: name, Port: port}
	}
	// copy to fts.table. Arrays are value types
	fts.table = new_tab
}

/* updates finger table to reflect new node i.e succ( [lo, hi) ) -> new_succ. Wrap arounds handled
Inclusive lo, exclusive hi
*/
func (fts *FTStruct) UpdateRange(lo [constants.ShaSize]byte, hi [constants.ShaSize]byte, new_succ *HostStruct) {
	for i := 0; i < len(fts.table); i++ {
		key := SSA.Add(fts.n, SSA.Pow2(uint32(i)))
		if InRangeHelp(key, lo, hi) {
			fts.table[i] = *new_succ // copy struct into array slot
		}
	}
}

/*
Initializer for a new FingerTable
n = the key/id for the local node
u_fn = function that maps key to a machine
*/
func New(n [constants.ShaSize]byte, u_fn UpdateFn) *FTStruct {
	ret := new(FTStruct)
	ret.n = n
	ret.Update(u_fn)
	return ret
}
