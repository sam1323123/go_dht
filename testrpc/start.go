package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

type Node struct {
	name     string
	port     string
	hostname string
	stop     bool
}

func (node *Node) F(arg *string, reply *bool) error {
	fmt.Printf("%s : %s\n", node.name, *arg)
	return nil
}

func (node *Node) Stop(arg *bool, reply *bool) error {
	node.stop = true
	return nil
}

/*
Does not Register !!!
Call this method to start the listener/service in a go routine
*/
func Start(node *Node) (net.Listener, error) {
	l, e := net.Listen("tcp", node.hostname+":"+node.port)
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

func main() {
	port := os.Args[1]
	name := os.Args[2]
	hostname := "127.0.0.1"
	node := new(Node)
	node.name = name
	node.hostname = hostname
	node.port = port
	node.stop = false
	rpc.Register(node)
	rpc.HandleHTTP()
	Start(node)
	for !node.stop {
		continue
	}

}
