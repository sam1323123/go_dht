package nodeapi

import (
	"fmt"
	SSA "go_dht/shasumarith"
	"testing"
)

/******** Helper Functions **********/

// test RPC find function helper
func testFind(hostname string, port string, t *testing.T) {
	key := SSA.FromInt(0)
	reply := HostData{Hostname: "", Port: ""}
	err := ConnectAndCall(hostname, port, "NAPI.Find", &key, &reply)
	if err != nil {
		t.Errorf("RPC error on Find. err = %s\n", err.Error())
	} else if reply.Port != port {
		t.Errorf("RPC error on Find. Got Port = %s\n", reply.Port)
	}
}

// connects to host machine and runc simple Hash Table ops
func testHashTableSimple(hostname string, port string, t *testing.T) {

	args := HTArgs{Key: "key", Value: "value"}
	reply := HTReply{Value: ""}
	err := ConnectAndCall(hostname, port, "NAPI.Put", &args, &reply)
	if err != nil {
		t.Errorf("RPC error = %s on Put to host = %s, port = %s \n", err.Error(), hostname, port)
	} else {
		fmt.Println("RPC Put success")
	}

	reply.Value = "" // reset Value
	err = ConnectAndCall(hostname, port, "NAPI.Get", &args, &reply)
	if err != nil {
		t.Errorf("RPC error = %s on Get \n", err.Error())
	} else if reply.Value != "value" { // no error wrong value
		t.Errorf("RPC error on Get. Got value = %s\n", reply.Value)
	}

	reply.Value = "" // reset value
	err = ConnectAndCall(hostname, port, "NAPI.Delete", &args, &reply)
	if err != nil {
		t.Errorf("RPC error = %s on Delete \n", err.Error())
	} else if reply.Value != "value" { // no error wrong value
		t.Errorf("RPC error on Delete. Got value = %s\n", reply.Value)
	}
}

/*************** Testing ***************/

// test if rpc works on a single node chord ring and simple hashtable functions
func TestRpcBasic(t *testing.T) {
	hostname := "localhost"
	port := "8080"
	ln := LocalInit(hostname, port, SSA.FromInt(0), nil)
	listener, err := NapiStart(ln)
	if err != nil {
		t.Errorf("Could not start RPC in TestRpcBasic\n")
		return
	}
	testHashTableSimple(hostname, port, t)
	testFind(hostname, port, t)
	NapiStop(listener) // stop the rpc service
}
