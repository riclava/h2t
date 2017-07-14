package handler

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// Service to expose
type Service struct {
	Name        string `json:"name"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

// ACLHandler of acl ops
type ACLHandler struct {
	ACL map[string]Service `json:"acl"`
}

// Put a new rule to map and do not affect opened connection
func (handler *ACLHandler) Put(name string, date string, description string) error {
	if handler.ACL == nil {
		handler.ACL = make(map[string]Service)
	}
	if _, found := handler.ACL[name]; found {
		log.Println("rule already exists [" + name + "], update it")
	}
	svc := Service{
		Name:        name,
		Date:        date,
		Description: description,
	}
	handler.ACL[name] = svc
	return nil
}

// Delete single rule from table. Doesn't affect opened connections
func (handler *ACLHandler) Delete(rule string) error {
	if _, found := handler.ACL[rule]; found {
		delete(handler.ACL, rule)
		return nil
	}
	return errors.New("No such rule [" + rule + "]")
}

// DeleteAll all acl rules, opened connections would not be affected
func (handler *ACLHandler) DeleteAll() {
	handler.ACL = make(map[string]Service)
}

// Copy makes a copy of acl
func (handler *ACLHandler) Copy() map[string]Service {
	res := make(map[string]Service)
	for k, v := range handler.ACL {
		res[k] = v
	}
	return res
}

// ServeHTTP - HTTP handler that accepts connection, check CONNECT method,
// find service and connect to tcp
func (handler *ACLHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		rw.Write([]byte("This is a http tunnel proxy, only CONNECT method is allowed."))
		return
	}

	host := req.URL.Host
	if _, ok := handler.ACL[host]; !ok {
		rw.Write([]byte("Your connection to [" + host + "] is not allowed."))
		log.Println("Connection from [", req.RemoteAddr, "] to [", host, "] is not allowed")
		return
	}

	hij, ok := rw.(http.Hijacker)
	if !ok {
		panic("HTTP Server does not support hijacking")
	}

	client, _, err := hij.Hijack()
	if err != nil {
		return
	}
	defer client.Close()

	log.Println("Connecting to host[", host, "]")
	server, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		log.Println("Connect to host[", host, "] failed, timeout")
		return
	}
	defer server.Close()

	client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))

	go io.Copy(server, client)
	io.Copy(client, server)

	log.Println("Connection closed by peer [", host, "]")
}
