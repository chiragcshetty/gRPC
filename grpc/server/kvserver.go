package main

import (
	"context"
	//"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"

	kvs "kvexample/kvstore/protocompiled/kvstore"
)

var (
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	port       = flag.Int("port", 8009, "The server port")
)

//*******************************************************************************************************************************
type storeServer struct {
	kvs.UnimplementedStoreServer
	// key is 24 bytes, but we are not enforcing it. 
	// Using a string instead (note: '[]byte' can't be used os map key in golang)
	store map[string]kvs.ValueTs 
	// TODO:Enforce key length (eg: map[[24]byte]kvs.ValueTs)
	//      But protobuff doesn't support fixed length byte array
	//		So need casting every time like: s.store[([24]byte)(key.Key[0:23])

	lock         sync.Mutex 
}
//*******************************************************************************************************************************


// Implementation of Get service
func (s *storeServer) Get(ctx context.Context, record *kvs.Record) (*kvs.ValueTs, error) {
	    //fmt.Println("Got a request for: ", record.String())
		s.lock.Lock()
		valuets, ok := s.store[record.GetKey()]
		s.lock.Unlock()
		if ok{
			return &valuets, nil
		}
	// Key was not found, return an unnamed ValueTS
	return &kvs.ValueTs{}, nil
}


// Implementation of Set service
func (s *storeServer) Set(ctx context.Context, record *kvs.Record) (*kvs.AckMsg, error) {
	//fmt.Println("\nGot a set for: ", record.String())

	affected_key  := record.GetKey()
	proposed_valuets := record.Valuets

	s.lock.Lock()
	current_valuets, ok := s.store[affected_key]

	if ok{
		//fmt.Printf("		Timestamps: Current: %d, Proposed: %d \n", current_valuets.GetTs(), proposed_valuets.GetTs())
		if proposed_valuets.GetTs() > current_valuets.GetTs(){
			s.store[affected_key] = *proposed_valuets
			//fmt.Printf("		Value replaced!\n\n")
		} else {
			//fmt.Printf("		Value not changed. No action needed\n\n")
		}	
	} else {
		s.store[affected_key] = *proposed_valuets
		//fmt.Printf("Key doesnt exist. Added\n\n")
	}
	s.lock.Unlock()

	// Server Acks in any case.
	return  &kvs.AckMsg{}, nil
}

//*******************************************************************************************************************************
// ************** Simran: template *************************
// Load 1M KVs: Fixed key size of 24 bytes and value size of 10bytes
// loadKV()
// ********************************************************

// loadFeatures some KV's
func (s *storeServer) loadKV() {
	for i := 1000; i < 10000; i++ {
		kv1 := kvs.Record{
			Key: fmt.Sprintf("%s%d", "KEY", i) ,
			Valuets:    &kvs.ValueTs{
				Value: fmt.Sprintf("%s%d", "VALUE", i),
				Ts: 102,
			},
		}
		s.store[kv1.GetKey()] = *kv1.Valuets
		printKV(&kv1)
	}	

	fmt.Printf("Setup Complete\n****************************------------------******************************************\n")
}


func printKV(kv *kvs.Record) {
	fmt.Printf("%s %s %d \n", kv.GetKey(), kv.Valuets.GetValue(), kv.Valuets.GetTs())
}


func newServer() *storeServer {
	s := &storeServer{ store: make(map[string]kvs.ValueTs) }
	s.loadKV()
	return s
}

//*******************************************************************************************************************************


func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	kvs.RegisterStoreServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}





//func serialize(kv *kvs.Record) string {
//	return fmt.Sprintf("%s %s %d", kv.KeyMsg.Key, kv.Valuets.Value, kv.ValueTs.Ts)
//}