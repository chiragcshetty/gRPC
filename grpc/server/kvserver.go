package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"os"
	"io/ioutil"

	"google.golang.org/grpc"

	// 'kvexample' is main module name (i.e was initialized as 'go mod init kvexample')
	// Compiled protobuff are stored at '/kvstore/protocompiled/kvstore'
	kvs "kvexample/kvstore/protocompiled/kvstore"
)

//########################################## Globals #######################################################################################
var (
	logType        = flag.String("log", "stdout", "Logging Options: off, file (stored at logs/server-port.txt), stdout (default)")
	port           = flag.String("port", "8009", "The server will listen on this port")
)

//########################################## The Actual KV store at this server ############################################################
type storeServer struct {
	kvs.UnimplementedStoreServer
	// key is 24 bytes, value is 10 bytes, but we are not enforcing it. 
	// Instaed we use a string for key, vlaue (note: '[]byte' can't be used as map key in golang)
	store map[string]kvs.ValueTs 
	// TODO:Enforce key length (eg: map[[24]byte]kvs.ValueTs)
	//      But protobuff doesn't support fixed length byte array
	//		So need casting every time like: s.store[([24]byte)(key.Key[0:23])

	// All read, write require locking of entire KV store. BAD!
	lock         sync.Mutex 
}

//####################################### gRPC Server Service implemmentation ###############################################################
// The Services are specified in kvstore/kvstore.proto

///************-------------------- GET Request Implementation -----------------------------**********************
// Request: GET (key). If key is found, send back the (value,ts). Else send ("", 0.0)
func (s *storeServer) Get(ctx context.Context, record *kvs.Record) (*kvs.ValueTs, error) {
	    log.Printf("!^! Got a Get request for: %s \n", record.GetKey())
		s.lock.Lock()
			valuets, ok := s.store[record.GetKey()]
		s.lock.Unlock()
		if ok{
			log.Printf("		Sending back: %s \n", valuets.String())
			return &valuets, nil
		}
	// Key was not found, return an empty valuets
	log.Printf("		Key not found!\n")
	return &kvs.ValueTs{}, nil
}


///************-------------------- SET Request Implementation -----------------------------************************
// Request: SET (key, propoosed_value, proposed_ts). If key doesn't exist, add it. In any case, Ack to the sender.
func (s *storeServer) Set(ctx context.Context, record *kvs.Record) (*kvs.AckMsg, error) {
	log.Printf("!~! Got a Set request for: %s\n", record.String())

	affected_key  := record.GetKey()
	proposed_valuets := record.Valuets

	s.lock.Lock()
		current_valuets, ok := s.store[affected_key]
		// Set the new vallue only if current ts < proposed ts. 
		// Else do nothing and just ack the request
		if ok {
				log.Printf("		Timestamps: Current: %f, Proposed: %f \n", current_valuets.GetTs(), proposed_valuets.GetTs())
				if proposed_valuets.GetTs() > current_valuets.GetTs(){
					s.store[affected_key] = *proposed_valuets
					log.Printf("		Value replaced!\n\n")
				} else {
					log.Printf("		Value not changed. No action needed\n\n")
				}	
		} else {  // If key is not found, add the key
				s.store[affected_key] = *proposed_valuets
				log.Printf("Key didn't exist. So, Added it. Thank me later!\n\n")
		}
	s.lock.Unlock()

	// Server Acks in any case.
	return  &kvs.AckMsg{}, nil
}

//#################################################### Utilities #############################################################################

// loadSome data into the KVstore to begin with
func (s *storeServer) loadKV() {
	for i := 1000; i < 10000; i++ {
		kv1 := kvs.Record{
			Key: fmt.Sprintf("%s%d", "KEY", i) ,
			Valuets:    &kvs.ValueTs{
				Value: fmt.Sprintf("%s%d", "VALUE", i),
				Ts: 1.0,
			},
		}
		s.store[kv1.GetKey()] = *kv1.Valuets
		printKV(&kv1)
	}	
	log.Printf("Setup Complete.\n\n")
}


func printKV(kv *kvs.Record) {
	log.Printf("(key, value, ts) = ( %s, %s, %f )\n", kv.GetKey(), kv.Valuets.GetValue(), kv.Valuets.GetTs())
}


func newServer() *storeServer {
	s := &storeServer{ store: make(map[string]kvs.ValueTs) }
	s.loadKV()
	return s
}


//####################################################################################################################################


func main() {
	flag.Parse()
	log.Printf("Starting. Logging Type %s, Port no %s ", *logType, *port ) 
	//******************************************** SETUP LOGGING **************************************************
	
	switch *logType {
	case "file":
			fmt.Sprintf("Logging type: %s, find logs at logs/server-%s.txt \n", *logType, *port) 
			file, err := os.OpenFile( ("./logs/server-" + *port + ".txt"), os.O_CREATE|os.O_WRONLY, 0666 )
			if err != nil { log.Printf("Failed to open log: ", err) }
			log.SetOutput(file); 
	case "stdout":
			log.SetOutput(os.Stdout)
	default: //off
			// source: https://stackoverflow.com/questions/10571182/how-to-disable-a-log-logger
			fmt.Sprintf("Logging switched off \n")
			log.SetFlags(0)
			log.SetOutput(ioutil.Discard) 
	}
	
	//***************************************** SETUP TCP SERVER********************************************************

	log.Printf("Replica Server Port No: %s \n\n", *port)
	lis, err := net.Listen("tcp", ":" + *port)
	if err != nil { log.Fatalf("Failed to listen: %v", err) }

	//***************************************** SETUP gRPC SERVER*******************************************************
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	kvs.RegisterStoreServer(grpcServer, newServer())

	//************************************* gRPC server now listening*****************************************************
	log.Printf("~~~~~~~~~ The server is now listening! ~~~~~~~~~~~~~~ \n\n")
	grpcServer.Serve(lis)

}