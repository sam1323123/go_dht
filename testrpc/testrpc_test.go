package testrpc

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"testing"
)

type NInterface interface {
	getName() string
	getPort() string
	getHostname() string
}

type Node struct {
	name     string
	port     string
	hostname string
}

func (node *Node) F(arg *string, reply *bool) error {
	fmt.Printf("%s : %s\n", node.name, *arg)
	return nil
}

func (node *Node) getName() string {
	return node.name
}

func (node *Node) getPort() string {
	return node.port
}

func (node *Node) getHostname() string {
	return node.hostname
}

type Node2 struct {
	name     string
	port     string
	hostname string
}

func (node *Node2) F2(arg *string, reply *bool) error {
	fmt.Printf("%s : %s\n", node.name, *arg)
	return nil
}

func (node *Node2) getName() string {
	return node.name
}

func (node *Node2) getPort() string {
	return node.port
}

func (node *Node2) getHostname() string {
	return node.hostname
}

/*
Does not Register !!!
Call this method to start the listener/service in a go routine
*/
func Start(node NInterface) (net.Listener, error) {
	l, e := net.Listen("tcp", node.getHostname()+":"+node.getPort())
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
func Stop(listener net.Listener) {
	fmt.Println("Shutting down RPC service ...")
	listener.Close()
	fmt.Println("RPC service successfully shutdown")
}

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

func TestRPC(t *testing.T) {
	host := "127.0.0.1"
	n1 := new(Node)
	n1.name = "n1"
	n1.hostname = host
	n1.port = "8080"

	n2 := new(Node2)
	n2.name = "n2"
	n2.hostname = host
	n2.port = "8081"

	rpc.Register(n1)
	rpc.Register(n2)
	rpc.HandleHTTP()

	ln1, err := Start(n1)
	if err != nil {
		t.Errorf("%s\n", err.Error())
		return
	}

	ln2, err := Start(n2)
	if err != nil {
		t.Errorf("%s\n", err.Error())
		return
	}
	var arg string = "Hi"
	var reply bool = false
	err = ConnectAndCall(n1.hostname, n1.port, "Node.F", &arg, &reply)
	if err != nil {
		t.Errorf("%s\n", err.Error())
		return
	}
	err = ConnectAndCall(n2.hostname, n2.port, "Node2.F2", &arg, &reply)
	if err != nil {
		t.Errorf("%s\n", err.Error())
		return
	}
	Stop(ln1)
	Stop(ln2)
}
