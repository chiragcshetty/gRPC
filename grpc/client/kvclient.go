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
	"io/ioutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// 'kvexample' is main module name (i.e was initialized as 'go mod init kvexample')
	// Compiled protobuff are stored at '/kvstore/protocompiled/kvstore'
	kvs "kvexample/kvstore/protocompiled/kvstore"
)

//############################################## Globals ###################################################################################
var (
	replicaClients []kvs.StoreClient   // One client per server replica
	clientid           = flag.Int("id", 0, "client id")
	numReplicas        = 0             // Number of server replicas. Obtained from serverAddrInfo.
	quorum             = 0             // floor((no of replicas)/2) + 1
	clientSignature float32            // eg: if clientid = 143, clientSignature is 0.143. 
									   // To have non-overlapping timestamps among clients
	// serverAddrInfo: file that contains replica info as "ip:port" - one per line
	serverAddrInfo     = flag.String("addrfile", "replicainfo.txt" , 
									"File containing server address in the format of host:port")
	logType        = flag.String("log", "stdout", `Logging Options: off, file 
									(stored at logs/server-port.txt), stdout (default)`)
)

//##################################### Get and Set from one replica ######################################################################

//------------------------------------------------- GET ------------------------------------------------------
// * Send Get request for 'key' to the replica handled by 'client'                 //
// * Request is sent as a kvs.Record(key, "", 0.0).                                //
// * On rcving the reply 'valuets' i.e (value, ts), forward it to main getKV()     //
// 	 through channel 'ch'. ch is 'chan<-' i.e you can send 'into' it               //
func getOneReplica(replicaNo int, client kvs.StoreClient, key string, 
											       ch chan<- *kvs.ValueTs  ) { 

	log.Printf(`ONE REPLICA %d - GET: Getting (value, ts) for 
				the key %s from replica no %d \n`, replicaNo, key, replicaNo)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kvRequest := kvs.Record{ Key: key ,}
	valuets, err := client.Get(ctx, &kvRequest)
	if err != nil {
		log.Printf(`WARNING! client.Get failed for the key %s 
					from replica no %d: %v \n`, key, replicaNo, err)
		return
	}
	log.Printf("ONE REPLICA %d - GET: Got - Key = %s, Value = %s, TS = %f \n", 
				replicaNo, key, valuets.GetValue(), valuets.GetTs())
	ch <- valuets
}


//------------------------------------------------- SET ------------------------------------------------------
// Send the SET(key, value, ts) request. Wait for an Ack and inform
// setKV() through 'ch'
func setOneReplica(replicaNo int,client kvs.StoreClient, key string, 
					         value string, ts float32, ch chan<- *kvs.AckMsg) {

	log.Printf(`ONE REPLICA %d - SET: Setting: ( key: %s, value : %s, ts : %f ) 
							at replica no %d \n`, replicaNo , key, value, ts, replicaNo)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kvRequest := kvs.Record{
		Key: key ,
		Valuets: &kvs.ValueTs{
		Value: value,
		Ts: ts,
		}, }
	ack, err := client.Set(ctx, &kvRequest)

	if err != nil {
		log.Printf(`WARNING! client.Set failed for the key %s at 
							 replica no %d: %v \n`, key, replicaNo, err)
		return
	}
	log.Printf(`ONE REPLICA %d - SET: Ack rcvd for key %s from replica 
										no %d \n`, replicaNo , key, replicaNo)
	ch <- ack
}

//##################################### Get and Set from entire replica cluster ############################################################

//------------------------------------------------- GET ------------------------------------------------------
// * Spawn a goroutine getOneReplica() per replica                           //
// * Create a channel to get replies from all getOneReplica()'s you spawn    //
// * Wait for quorum replies. Get max timestamp among them, return it        //
//   and the corresponding value                                             //
func getKV(key string) (float32, string) {

	//buffer size = numReplicas (else ch<- block at sending end)
	getChannel := make(chan *kvs.ValueTs, numReplicas) 
	numReplies := 0
	var maxTs float32 = 0.0
	value      := ""

	for i:=0; i <numReplicas; i++ { // launch a handler per replica
		go getOneReplica(i, replicaClients[i], key, getChannel) }

	log.Printf(`Launched GET handlers for all replicas 
				for Key %s \n`,key )

	// loop until quorum number of replies
	for ( numReplies < quorum ) {
		// non-blocking wait for reads from getChannel
		select {
			case replyValuets := <-getChannel:
				replicaTs := replyValuets.GetTs()
				// If key was not found, replicaTS will be 0.0 and 
				// maxTs will end up being 0.0, value = ""

				if  (replicaTs > maxTs) { 
					maxTs = replicaTs 
					value = replyValuets.GetValue()  
				}
				numReplies = numReplies + 1
				log.Printf("GET Key %s : Number of Replies: %d \n",
							 key, numReplies)

			default:
				time.Sleep(20 * time.Nanosecond)
		}
	}
	log.Printf("Key %s : getKV() done : maxTS = %f, value = %s \n",
						 key, maxTs, value)
	return maxTs, value
}


//------------------------------------------------- SET ------------------------------------------------------
// Send Set requests to all replicas and wait for quorum acks 
func setKV(key string, value string, ts float32) int {
	//buffer size = numReplicas (else ch<- block at sending end)
	ackChannel := make(chan *kvs.AckMsg, numReplicas) 
	numAcks    := 0

	for i:=0; i <numReplicas; i++ { // launch a handler per replica
		go setOneReplica(i, replicaClients[i], key, value, ts, ackChannel)
	}

	log.Printf(`Launched SET handlers for all replicas 
				for Key %s, Value %s, TS %f \n`,key, value, ts )

	for ( numAcks < quorum ) {
		select{
			case _ = <-ackChannel:
				numAcks = numAcks + 1
				log.Printf("SET Key %s : Number of Acks: %d \n",
							 key, numAcks)
			default:
				time.Sleep(20 * time.Nanosecond)
		}
	}
	log.Printf("Key %s : setKV() done \n", key)
	return numAcks
}

//######################################### MAIN Methods: read() & write() ############################################################

// Reliably read 'key' from the KV store
func read(key string) string {
	// GET Phase
	maxTs, value := getKV(key)
	// Compute a higher TS than all observed
	tsNew := float32(math.Floor(float64(maxTs+1))) + clientSignature
	// SET Phase ('Ensures future reads will see what I read')
	setKV(key, value, tsNew)
	log.Printf(`READ for Key %s completed. Value = %s, 
							ts = %f \n`, key, value, tsNew)
	return value
}

// Reliably write (key, newValue) to the KV store
// If key doesn't exist, it will be added.
func write(key string, newValue string) {
	// GET Phase (to get max timestamp)
	maxTs, _ := getKV(key)
	// Compute a higher TS than all observed
	tsNew := float32(math.Floor(float64(maxTs+1))) + clientSignature
	// SET Phase
	setKV(key, newValue, tsNew)
	log.Printf(`WRITE to Key %s completed. Value = %s, 
							ts = %f \n`, key, newValue, tsNew)
	return 
}

//#################################################### Utilities ############################################################
func logKV(key string, valuets *kvs.ValueTs) {
	log.Printf("	GOT: Key = %s, Value = %s, TS = %f \n", 
					key, valuets.GetValue(), valuets.GetTs())
}

//###################################################### MAIN ###############################################################
func main() {

	flag.Parse()

	log.Printf("Starting. Logging Type %s, client ID %d ", *logType, *clientid ) 
	//******************************************** SETUP LOGGING **************************************************
	switch *logType {
	case "file":
			fmt.Sprintf("Logging type: %s, find logs at logs/client-%d.txt \n", *logType, *clientid) 
			file, err := os.OpenFile( fmt.Sprintf("./logs/client-%d.txt", *clientid), os.O_CREATE|os.O_WRONLY, 0666 )
			if err != nil { log.Printf("Failed to open log: ", err) }
			log.SetOutput(file); 
	case "stdout":
			log.SetOutput(os.Stdout)
	default: //off
			// source: https://stackoverflow.com/questions/10571182/how-to-disable-a-log-logger
			log.Printf("Logging switched off \n")
			log.SetFlags(0)
			log.SetOutput(ioutil.Discard) 
	}

	//************************************* SETUP: one client per replica *************************************************
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//--------------------------- Read replicas' info from serverAddrInfo file --------------------------
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
			log.Printf("WARNING! Failed to dial %s : %v", replicaddr, err)
		}
		defer conn.Close()
		replicaClients = append(replicaClients, kvs.NewStoreClient(conn))
		numReplicas = numReplicas + 1 
    }

	if err := scanner.Err(); err != nil {log.Fatal(err)}

	// quorum = floor((no of replicas)/2) + 1 (i.e simple majority)
	quorum = int(math.Floor(float64(numReplicas/2))) + 1
	log.Printf("No. of Replicas found = %d & quorum = %d \n\n", numReplicas, quorum)

	sig, _    := strconv.ParseFloat( fmt.Sprintf("0.%d",*clientid ), 3)
	clientSignature = float32(sig)


	//************************************* Test Workload *************************************************

	var i int32 = 0; value := ""; key := ""

	start := time.Now()
	for i=200; i<10000; i++{
		key = fmt.Sprintf("KEY%d%d",700 ,i%10)
		value = read(key)
		log.Printf("Recieved for key %s: %s \n\n", key, value)

		newValue :=  fmt.Sprintf("VAL%d%d",*clientid,i)
		write(key, newValue)
		log.Printf("Written for key %s: %s \n", key, newValue)
		log.Printf("***************************************************************\n\n")
	}
	duration := time.Since(start)
	fmt.Println(duration)

}