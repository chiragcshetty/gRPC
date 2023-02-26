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
	clientid           = flag.Int("id", 0, "client id")
	serverAddr         = flag.String("addr", "localhost:8009", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.example.com", "The server name used to verify the hostname returned by the TLS handshake")
)

func printKV(key string, valuets *kvs.ValueTs) {
	fmt.Printf("GOT: %s %s %d \n", key, valuets.GetValue(), valuets.GetTs())
}


// gets value for the given key
func getKV(client kvs.StoreClient, key string) {
	//fmt.Printf("Getting value and ts for the key %s \n", key)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kv1 := kvs.Record{
		Key: key ,
	}
	_, err := client.Get(ctx, &kv1)
	if err != nil {
		log.Fatalf("client.Get failed: %v", err)
	}
	//printKV(key, valuets)
}

func setKV(client kvs.StoreClient, key string, value string, ts int32) {
	//fmt.Printf("Setting: ( key: %s, value : %s, ts : %d ) \n", key, value, ts)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kv1 := kvs.Record{
		Key: key ,
		Valuets:    &kvs.ValueTs{
		Value: value,
		Ts: ts,
		},
	}
	_, err := client.Set(ctx, &kv1)
	if err != nil {
		log.Fatalf("client.Get failed: %v", err)
	}
	//fmt.Println("Ack rcvd:", ack.String())
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
	fmt.Printf("Client id %d \n", *clientid)
	var opts []grpc.DialOption
	
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))


	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := kvs.NewStoreClient(conn)

	// Looking for a valid feature
	var i int32 = 0

	start := time.Now()
	for i=200; i<20000; i++{
		setKV(client, "KEY9991", fmt.Sprintf("%s-%d-%d", "Val", *clientid ,i), i)
		setKV(client, "KEY99910", fmt.Sprintf("%s-%d-%d", "Val2", *clientid, i), i)
		getKV(client, "KEY9991")
		getKV(client, "KEY99910")
	}
	duration := time.Since(start)
	fmt.Println(duration)

}