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

type storeServer struct {
	kvs.UnimplementedStoreServer
	// key is 24 bytes, but we are not enforcing it 
	store map[[24]byte]kvs.ValueTs 
	// TODO:  Is it a good idea to use proto object as key? 
	// TODO: Enforce key length (eg: map[[24]byte]kvs.ValueTs)
	//       But protobuff doesn't support fixed length byte array
	//		So need casting every time like: s.store[([24]byte)(key.Key[0:23])]

	mu         sync.Mutex 
}

// GetFeature returns the feature at the given point.
func (s *storeServer) Get(ctx context.Context, key *kvs.Key) (*kvs.ValueTs, error) {
		val_ts, ok := s.store[([24]byte)(key.Key[0:23])]
		if ok{
			return &val_ts, nil
		}

	// No feature was found, return an unnamed feature
	return &kvs.ValueTs{}, nil
}




// loadFeatures loads features from a JSON file.
func (s *storeServer) loadKV() {

	kv1 := kvs.Record{
		Key: &kvs.Key{ Key: []byte{1,2,3,4,5,6,7,8,9,10,1,2,3,4,5,6,7,8,9,10,1,2,3,4}},
		Valuets:    &kvs.ValueTs{
			Value: []byte{1,2,3,4,5,6,7,8,9,10},
			Ts: 102,
		},
	}
	fmt.Println("TS:", kv1.Valuets.Ts)
	s.store[([24]byte)(kv1.Key.Key)] = *kv1.Valuets
	fmt.Println("TS:", kv1.Valuets.Ts)
}



//func serialize(point *pb.Point) string {
//	return fmt.Sprintf("%d %d", point.Latitude, point.Longitude)
//}

func newServer() *storeServer {
	s := &storeServer{store: make(map[[24]byte]kvs.ValueTs )}
	s.loadKV()
	return s
}

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

// exampleData is a copy of testdata/route_guide_db.json. It's to avoid
// specifying file path with `go run`.
var exampleData = []byte(`[{
    "key": {
        "key": [10,20,34,45]
    },
    "valuets":{
		"value":"VALUE1",
		"ts": 100
	}
}]`)