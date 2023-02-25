package main

import (
	"context"
	"flag"
	"log"
	"time"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	kvs "kvexample/kvstore/protocompiled/kvstore"
)

var (
	serverAddr         = flag.String("addr", "localhost:8009", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.example.com", "The server name used to verify the hostname returned by the TLS handshake")
)

func printKV(key string, valuets *kvs.ValueTs) {
	fmt.Printf("GOT: %s %s %d \n", key, valuets.GetValue(), valuets.GetTs())
}


// printFeature gets the feature for the given point.
func getKV(client kvs.StoreClient, key string) {
	fmt.Printf("Getting value and ts for the key %s \n", key)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kv1 := kvs.Record{
		Key: key ,
	}
	valuets, err := client.Get(ctx, &kv1)
	if err != nil {
		log.Fatalf("client.Get failed: %v", err)
	}
	printKV(key, valuets)
}

// ************** Simran: template *************************
// func clientGet(key string){}
// func clientSet(key string, value string){}
// func workloadGenration()
// Correctness check
// etcd compare
// ************** Simran: template *************************

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))


	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := kvs.NewStoreClient(conn)

	// Looking for a valid feature
	getKV(client, "KEY9991")

}