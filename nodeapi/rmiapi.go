package nodeapi

import (
	"crypto/sha1"
	"fmt"
	CM "go_dht/chordmap"
	K "go_dht/constants"
	FT "go_dht/fingertable"
	SSA "go_dht/shasumarith"
	"net"
	"net/http"
	"net/rpc"
)

/*
Library to handle RMI calls to this node
*/

// Exported struct used to contain arguments for Hash Table functions like Put/Get/Delete for use in RMI calls
type HTArgs struct {
	Key   string // exported. key for the hash table
	Value string // exported. value for the hash table
}

// Exported struct used to contain arguments for Hash Table functions like Put/Get/Delete for use in RMI calls
type HTReply struct {
	Value string
}

// struct to package information about a host machine for transmission using RPCs
type HostData struct {
	Hostname, Port string // Exported
}

//types and  args struct for notifyPred
type jevent int

const (
	jeventJoining jevent = 0 // node is joining
	jeventJoined  jevent = 1 // node has joined
)

type joinNotice struct {
	event  jevent
	caller *HostData // can be used for checking if caller is the successor
}

type JoinRequest struct {
	key  [K.ShaSize]byte
	conn *HostData
}

/*********** Helper Functions ****************/

// Convenience method to make rpc calls to srv_addr:srv_port using method
func ConnectAndCall(srv_addr string, srv_port string, method string, args interface{}, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", srv_addr+":"+srv_port)
	if err != nil {
		return err
	}
	err = client.Call(method, args, reply) // make rpc call
	if err != nil {
		return err
	}
	return nil // no error
}

/************ End Helper *************/

// Node api struct to contain methods for use in RMI Register method
type NAPI struct {
	ln *LocNodeStruct // information about the local node. Must not be nil
}

/******** RMI Methods for NAPIStruct **********/

/*
Hash Table get method used by client. Assumes ln != nil
*/
func (napi *NAPI) Get(args *HTArgs, reply *HTReply) error {
	ln := napi.ln
	shakey := sha1.Sum([]byte(args.Key))
	if ln.StoresKey(shakey) { // store locally and return the error
		val, err := ln.cm.Get(args.Key)
		reply.Value = val
		return err
	} else { // must find in chord ring
		srv_addr, srv_port, err := ln.ft.Find(shakey) // seearch in fingertable
		if err != nil {
			return err
		}
		return ConnectAndCall(srv_addr, srv_port, "NAPI.Get", args, reply)
	}
}

/*
Hash Table Put method used by client. reply is overwritten to containe empty string
*/
func (napi *NAPI) Put(args *HTArgs, reply *HTReply) error {
	ln := napi.ln
	shakey := sha1.Sum([]byte(args.Key))
	if ln.StoresKey(shakey) { // store locally and return the error
		err := ln.cm.Put(args.Key, args.Value)
		reply.Value = ""
		return err
	} else { // must find in chord ring
		srv_addr, srv_port, err := ln.ft.Find(shakey) // seearch in fingertable
		if err != nil {
			return err
		}
		return ConnectAndCall(srv_addr, srv_port, "NAPI.Put", args, reply)
	}
}

/*
Hash Table Delete method used by client
*/
func (napi *NAPI) Delete(args *HTArgs, reply *HTReply) error {
	ln := napi.ln
	shakey := sha1.Sum([]byte(args.Key))
	if ln.StoresKey(shakey) { // store locally and return the error
		val, err := ln.cm.Delete(args.Key)
		reply.Value = val
		return err
	} else { // must find in chord ring
		srv_addr, srv_port, err := ln.ft.Find(shakey) // seearch in fingertable
		if err != nil {
			return err
		}
		return ConnectAndCall(srv_addr, srv_port, "NAPI.Delete", args, reply)
	}
}

/*retrieve the id/sha-key associated with this node. Assumes ln != nil
args is not used
*/
func (napi *NAPI) GetN(args *[K.ShaSize]byte, reply *[K.ShaSize]byte) error {
	ln := napi.ln
	if ln.pred == nil { // single node chord ring
		*reply = SSA.MaxVal() // entire key space
		return nil
	}
	*reply = ln.pred_end
	return nil
}

/*Find the node that is in charge of key
 */
func (napi *NAPI) Find(key *[K.ShaSize]byte, reply *HostData) error {
	ln := napi.ln
	if ln.StoresKey(*key) {
		reply.Hostname = ln.hostname
		reply.Port = ln.port
		return nil
	}
	// find in Chord ring
	srv_addr, srv_port, err := ln.ft.Find(*key) // seearch in fingertable
	if err != nil {
		return err
	}
	return ConnectAndCall(srv_addr, srv_port, "NAPI.Find", key, reply)
}

/*
Unexposed function for internal api use
CallerError if notifyPred was not invoked by this node's successor
Else returns the error thrown by ln.SetState
reply is unsued
*/
func (napi *NAPI) notifyPred(notice *joinNotice, reply *bool) error {
	ln := napi.ln
	succ_ip, succ_port, err := ln.ft.Find(SSA.Add(ln.end, SSA.FromInt(1))) // get successor
	if err != nil {
		return err
	}
	if (notice.caller.Hostname != succ_ip) || (notice.caller.Port != succ_port) {
		// if not invoked by succ
		return NewNapiCallerError()
	}
	var new_state nodestate
	if notice.event == jeventJoining {
		new_state = Busy
	} else if notice.event == jeventJoined {
		new_state = Free
		// update fingertable
		ln.ft.UpdateRange(SSA.Add(SSA.FromInt(1), ln.end), SSA.Add(SSA.FromInt(1), ln.joiner.N), &FT.HostStruct{Hostname: ln.joiner.Conn.Hostname,
			Port: ln.joiner.Conn.Port})
	}
	return ln.SetState(new_state)

}

/*
Unexposed
Submodule of RegisterJoin. Must be called on succ(key) to ensure correctness. No error if can join else JoinError. Else returns first caught error
Ensures both succ and pred are not locked, then initialize the loc node struct and returns it to
caller for use.
*/
func (napi *NAPI) registerJoinSucc(request *JoinRequest, reply **LocNodeStruct) error {
	ln := napi.ln
	var pred *HostData = ln.pred
	var err error = ln.SetState(BusyJoin)
	if err != nil { // local node is busy
		return err
	}
	var jln *LocNodeStruct
	var jcm *CM.ChordMapStruct
	if pred == nil { // single node chord ring, skip predecessor stages
		// joiner's predecessor is also its successor
		jln, err = LocalInit(request.conn.Hostname, request.conn.Port, request.key, &HostData{Hostname: ln.hostname, Port: ln.port}) // joiners local node struct
		if err != nil {
			ln.SetState(Free) // release local node
			return err
		}
		jcm, err = ln.cm.PartitionTable(SSA.Add(request.key, SSA.FromInt(1)))
	} else { //not single node chord ring i.e pred != nil
		args := joinNotice{event: jeventJoining, caller: &HostData{Hostname: ln.hostname, Port: ln.port}}
		reply := true
		err := ConnectAndCall(pred.Hostname, pred.Port, "NAPI.notifyPred", &args, &reply)
		if err != nil {
			ln.SetState(Free) // release local node
			return err        // error with setting state or wrong predecessor
		}
		jln, err = LocalInit(request.conn.Hostname, request.conn.Port, request.key, pred)
		if err != nil {
			ln.SetState(Free) // release local node
			return err
		}
		jcm, err = ln.cm.PartitionTable(SSA.Add(request.key, SSA.FromInt(1)))
	}
	// setup joiner struct, fill the jln and fill the reply value
	if err != nil {
		ln.SetState(Free) // release the node
		return err
	}
	joiner := Joiner{
		Table: jcm,
		N:     request.key,
		Conn:  request.conn}
	ln.joiner = &joiner // set the joiner container
	jln.cm = jcm        // set the cm for the reply
	*reply = jln
	return nil
}

/*
Helper method for Joined
Must be called on succ(request.key)
Requires Joiner to complete setup and RegisterJoin to be previously called successfully. Alerts pred and succ of new node
Does not try until success. Returns err on first error caught
Fingertables of only the succ and pred(done in notifyPred) are updated.
*/
func (napi *NAPI) joinedSucc(request *JoinRequest, reply *bool) error {
	ln := napi.ln
	var err error = nil
	if ln.pred != nil { // alert predecessor
		jn := joinNotice{event: jeventJoined, caller: &HostData{Hostname: ln.hostname, Port: ln.port}}
		err = ConnectAndCall(ln.pred.Hostname, ln.pred.Port, "NAPI.notifyPred", &jn, reply)
		if err != nil { // if error raised by pred, then abort atomic transaction
			return err
		}
	}
	// update fingertable, change pred, clear joiner, set state
	ln.ft.UpdateRange(SSA.Add(SSA.FromInt(1), ln.pred_end), SSA.Add(SSA.FromInt(1), ln.end), &FT.HostStruct{Hostname: request.conn.Hostname,
		Port: request.conn.Port})
	*(ln.pred) = *(request.conn) // copy the struct
	ln.pred_end = request.key
	ln.joiner = nil
	err = ln.SetState(Free) // should never return an err
	return nil
}

/*
Can be invoked on any node
Success message on successful joining by new node after calling registerJoinedSucc
Requires Joiner to complete setup and RegisterJoin to be previously called successfully. Alerts pred and succ of new node
Does not try until success. Returns err on first error caught
Fingertables are not updated here. Fingertables are periodically refreshed
*/
func (napi *NAPI) Joined(request *JoinRequest, reply *bool) error {
	var succ HostData
	err := napi.Find(&(request.key), &succ)
	if err != nil {
		return err
	}
	err = napi.joinedSucc(request, reply)
	return err
}

/*
request: contains joiner's key and ip info
reply: necessary info for init. Untouched if error raised
Called by a joiner wanting id = key. No error if can join else JoinError. Else returns first caught error
First finds succ, then ensures both are not locked, then initialize the loc node struct and returns it to
client for use.
*/
func (napi *NAPI) RegisterJoin(request *JoinRequest, reply **LocNodeStruct) error {
	var succ HostData
	err := napi.Find(&(request.key), &succ)
	if err != nil {
		return err
	}
	err = ConnectAndCall(succ.Hostname, succ.Port, "NAPI.registerJoinSucc", request, reply)
	return err
}

/********* RMI end *************/

/*
Call this method to register the rpc service and start the listener/service in a go routine
*/
func NapiStart(loc_node *LocNodeStruct) (net.Listener, error) {
	napi := new(NAPI)
	napi.ln = loc_node
	rpc.Register(napi)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", loc_node.hostname+":"+loc_node.port)
	if e != nil {
		// detected error
		fmt.Printf("Cannot start RPC service. %s \n", e.Error())
		return l, e
	}
	go http.Serve(l, nil) // accepts connections on listener l and handles them using default handler dispatcher
	fmt.Println("RPC service started successfully")
	return l, nil
}

/*
Call this method to stop the rpc service
*/
func NapiStop(listener net.Listener) {
	fmt.Println("Shutting down RPC service ...")
	listener.Close()
	fmt.Println("RPC service successfully shutdown")
}
