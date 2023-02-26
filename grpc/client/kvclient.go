package main

import (
	"context"
	"flag"
	"log"
	"time"
	"fmt"
	"os"
	"bufio"
	"strconv"
	"math"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	kvs "kvexample/kvstore/protocompiled/kvstore"
)

// "localhost:8009"
var (
	replicaClients []kvs.StoreClient
	clientid           = flag.Int("id", 0, "client id")
	numReplicas        = 0 // number of server replicas
	quorum             = 0
	clientSignature float32
	serverAddrInfo     = flag.String("addrfile", "replicainfo.txt" , "File containing server address in the format of host:port")
)

func printKV(key string, valuets *kvs.ValueTs) {
	fmt.Printf("GOT: %s %s %f \n", key, valuets.GetValue(), valuets.GetTs())
}

//************************************************************************************************************************
// gets value for the given key
func getOneReplica(client kvs.StoreClient, key string, ch chan<- *kvs.ValueTs  ) { //ch: channel sendint end
	fmt.Printf("Getting value and ts for the key %s \n", key)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kv1 := kvs.Record{
		Key: key ,
	}
	valuets, err := client.Get(ctx, &kv1)
	if err != nil {
		//log.Fatalf("client.Get failed: %v", err)
		return
	}
	printKV(key, valuets)
	ch <- valuets
}

func setOneReplica(client kvs.StoreClient, key string, value string, ts float32, ch chan<- *kvs.AckMsg) {
	fmt.Printf("Setting: ( key: %s, value : %s, ts : %f ) \n", key, value, ts)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kv1 := kvs.Record{
		Key: key ,
		Valuets:    &kvs.ValueTs{
		Value: value,
		Ts: ts,
		},
	}
	ack, err := client.Set(ctx, &kv1)
	if err != nil {
		//log.Fatalf("client.Get failed: %v", err)
		return
	}
	fmt.Println("Ack rcvd:", ack.String())
	ch <- ack
}

//************************************************************************************************************************
func getKV(key string) (float32, string) {
	getChannel := make(chan *kvs.ValueTs, numReplicas)  // buffer size = numReplicas
	numReplies := 0
	var maxTs float32 = 0.0
	value      := ""
	for i:=0; i <numReplicas; i++ {
		go getOneReplica(replicaClients[i], key, getChannel)
	}

	for (numReplies<quorum){
		select {
			case replyValuets := <-getChannel:
				replicaTs := replyValuets.GetTs() // If key was not found, replicaTS will be 0.0 and maxTs will end up being 0.0, value = ""
				if  (replicaTs > maxTs) { 
					maxTs = replicaTs 
					value = replyValuets.GetValue()  
				}
				numReplies = numReplies + 1
				fmt.Printf("Num replies : %d\n", numReplies)
			default:
				time.Sleep(20 * time.Nanosecond)
		}
	}
	fmt.Printf("getKV Done \n")
	return maxTs, value
}

func setKV(key string, value string, ts float32) int {
	ackChannel := make(chan *kvs.AckMsg, numReplicas)  // buffer size = numReplicas
	numAcks  := 0
	for i:=0; i <numReplicas; i++ {
		go setOneReplica(replicaClients[i], key, value, ts, ackChannel)
	}

	for (numAcks<quorum) {
		select{
			case _ = <-ackChannel:
				numAcks = numAcks + 1
			default:
				time.Sleep(20 * time.Nanosecond)
		}
	}
	fmt.Printf("setKV Done \n")
	return numAcks
}

//************************************************************************************************************************
func read(key string) string {
	maxTs, value := getKV(key)
	tsNew := float32(math.Floor(float64(maxTs+1))) + clientSignature
	setKV(key, value, tsNew)
	return value
}

func write(key string, newValue string) {
	maxTs, _ := getKV(key)
	tsNew := float32(math.Floor(float64(maxTs+1))) + clientSignature
	setKV(key, newValue, tsNew)
	return 
}


//************************************************************************************************************************
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

	
	//****************************************************************************
	// source: https://stackoverflow.com/questions/8757389/reading-a-file-line-by-line-in-go
	file, err := os.Open(*serverAddrInfo)
	if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

	scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        replicaddr := scanner.Text()
		conn, err := grpc.Dial(replicaddr, opts...)
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		defer conn.Close()
		replicaClients = append(replicaClients, kvs.NewStoreClient(conn))
		numReplicas = numReplicas + 1 
    }

	if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

	quorum = int(math.Floor(float64(numReplicas/2))) + 1
	fmt.Printf("No. of Replicas = %d & quorum = %d \n\n", numReplicas, quorum)

	sig, _    := strconv.ParseFloat( fmt.Sprintf("0.%d",*clientid ), 3)
	clientSignature = float32(sig)
	//*******************************************************************************

	var i int32 = 0
	value := ""
	key := ""

	start := time.Now()
	for i=200; i<10000; i++{
		fmt.Println(i, " ---------------------------------------------------------------------------------------------------")
		key = fmt.Sprintf("KEY%d%d",700 ,i%10)
		value = read(key)
		fmt.Printf("Recieved for key %s: %s \n\n", key, value)

		newValue :=  fmt.Sprintf("VAL%d%d",*clientid,i)
		write(key, newValue)
		fmt.Printf("Written for key %s: %s \n", key, newValue)
		fmt.Printf("***************************************************************\n\n")
	}
	duration := time.Since(start)
	fmt.Println(duration)

}