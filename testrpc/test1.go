package main

import (
	"fmt"
	"net/rpc"
)

type Node struct {
	name     string
	port     string
	hostname string
}

func (node *Node) F(arg *string, reply *bool) error {
	fmt.Printf("%s : %s\n", node.name, *arg)
	return nil
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
	p1, p2 := "8080", "8081"
	hostname := "127.0.0.1"
	arg := "Hi"
	reply := false
	ConnectAndCall(hostname, p1, "Node.F", &arg, &reply)
	ConnectAndCall(hostname, p2, "Node.F", &arg, &reply)
	ConnectAndCall(hostname, p1, "Node.Stop", &reply, &reply)
	ConnectAndCall(hostname, p2, "Node.Stop", &reply, &reply)
}
