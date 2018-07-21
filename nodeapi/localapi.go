package nodeapi

import (
	CM "go_dht/chordmap"
	K "go_dht/constants"
	FT "go_dht/fingertable"
	SSA "go_dht/shasumarith"
	"sync"
)

type nodestate int

const (
	Free     nodestate = 0
	BusyJoin nodestate = 1 // in the process of registering a joiner
	Busy     nodestate = 2 // general busy case
)

// struct used as data storage container for the joining processs
type Joiner struct {
	//Pred, Succ [K.ShaSize]byte    // the id/key of the successor and predecessor of the Joined node
	Table *CM.ChordMapStruct // the table filled with the (key,values) tha joiner will take ownership
	N     [K.ShaSize]byte
	Conn  *HostData
}

/* class containing data for a local node
Note: Only do ft lookups after check on local node storage. Ft must also be updated if successor changes. To be safe both pred/succ change should update
ft.
*/
type LocNodeStruct struct {
	hostname, port string
	end            [K.ShaSize]byte    // id for this node. Inclusive
	pred           *HostData          // nil if chord ring has only this node in it. HostData defined in rmiapi.go
	pred_end       [K.ShaSize]byte    // the id/end of the predecessor node i.e the last key the predecessor is in charge of. Must be set if pred != nil
	ft             *FT.FTStruct       // fingertable. Is set correctly even for single node chord ring to ensure successor lookups are correct
	cm             *CM.ChordMapStruct // local hash table
	state          nodestate          // TODO: current state of the local node
	state_lock     *sync.Mutex
	joiner         *Joiner // Set to a Joiner struct if state == BusyJoin else should be nil
}

/********* Helper functions **********************/

/* Checks if x is within (start, end] i.e exclusive start, inclusive end. Also handles case of wrap around where end < start.
If end == start, returns false.
*/
func InRangeHelp(x [K.ShaSize]byte, start [K.ShaSize]byte, end [K.ShaSize]byte) bool {
	if SSA.Cmp(start, end) == SSA.Equal {
		return false
	} else if SSA.Cmp(start, end) == SSA.Less { // no wrap around
		return SSA.InRange(x, SSA.Add(start, SSA.FromInt(1)), end) // start+1 <= x <= end
	} else { // wrap around case
		return InRangeHelp(x, start, SSA.MaxVal()) || SSA.InRange(x, SSA.FromInt(0), end)
		// x in (start, MaxVal] || x == 0 || x in [0,end]
	}
}

/*********** Methods for LocNode Struct *************/

/*
True if key is under charge of this node. i.e within range of (pred_end, end] if pred != nil. Else return true
*/
func (lns *LocNodeStruct) StoresKey(key [K.ShaSize]byte) bool {
	if lns.pred == nil {
		return true // only 1 node in chord ring so Loc Node stores key
	}
	return InRangeHelp(key, lns.pred_end, lns.end)
}

/*
True if has predecessor. Checks if pred == nil. No predecessor implies a single node chord ring.
*/
func (lns *LocNodeStruct) HasPred() bool {
	return lns.pred != nil
}

/*
Checks if the local node can accept a join request with the given key.
Returns true if the local node is in currently in charge of key and is not undergoing another join process
*/
func (lns *LocNodeStruct) CanJoin(key [K.ShaSize]byte) bool {
	return (lns.state != BusyJoin) && lns.StoresKey(key)
}

/*
Call is synchronized
Tries to set lns.state to new_state. If node busy and new_state != Free then return Busy error.
Else if new_state is free and node is busy || node is free then OK
*/
func (lns *LocNodeStruct) SetState(new_state nodestate) error {
	lns.state_lock.Lock() // synchronize as rmi's are concurrent
	if new_state == Free || lns.state == Free {
		lns.state = new_state
		return nil
	} else { // lns.state andnew_state are both busy types
		return NewNapiBusyError()
	}
	lns.state_lock.Unlock()
	return nil
}

/*
Getter method for node state
*/
func (lns *LocNodeStruct) GetState() nodestate {
	return lns.state
}

/*
TODO: ELse case i.e more than 1 node in ring
Initializes the local node by creating the local node struct. ChordMap is empty on init. FingerTable is filled using Napi.Find() i.e reflects
the current state of the DHT
hostname: the public ip of the local node on which RPC is run
port:  the port on which rpc is run
end = id or last key for the local node. If pred == nil, does not matter
pred = Info of the predecessor machine. If != nil then LocalInit will contact the machine for pred_end info. If == nil then function assumed there
must be only 1 machine in the chord ring
*/
func LocalInit(hostname string, port string, end [K.ShaSize]byte, pred *HostData) (*LocNodeStruct, error) {
	ret := new(LocNodeStruct) // ret is a pointer
	ret.hostname = hostname
	ret.port = port
	ret.state_lock = &sync.Mutex{}
	ret.state = Free
	ret.joiner = nil
	if pred == nil {
		ret.end = SSA.MaxVal()
		ret.pred = nil
		ret.ft = FT.New(end, func(k [K.ShaSize]byte) (string, string) { return hostname, port })
		ret.cm = CM.New(SSA.FromInt(0), SSA.FromInt(0)) // entire hash space
	} else {
		ret.pred = &HostData{Hostname: pred.Hostname, Port: pred.Port}                    // do a copy
		err = ConnectAndCall(pred.Hostname, pred.Port, "NAPI.GetN", nil, &(ret.pred_end)) // get pred_end. Args is not used
		if err != nil {
			return nil, err
		}
		ret.ft = FT.New(end, func(key [K.ShaSize]byte) {
			var ret HostData
			ConnectAndCall(pred.Hostname, pred.Port, "Napi.Find", &key, &ret) // TODO: error is not caught during failure
			return ret.Hostname, ret.Port
		})
		ret.cm = CM.New(SSA.Add(SSA.FromInt(1), ret.pred_end), SSA.Add(SSA.FromInt(1), end)) // [start, end)
	}
	return ret, nil
}
