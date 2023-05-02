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
	//"bufio"
	//"strings"
	"io"
	"os/exec"
	"time"

	"google.golang.org/grpc"

	//'kvexample' is main module name (i.e was initialized as 'go mod init kvexample')
	// Compiled protobuff are stored at '/kvstore/protocompiled/kvstore'
	kvs "kvexample/kvstore/protocompiled/kvstore_file"
)

//########################################## Globals #######################################################################################
var (
	logType        = flag.String("log", "off", "Logging Options: off (default), file (stored at logs/server-port.txt), stdout ")
	port           = flag.String("port", "8009", "The server will listen on this port")
	//filesize       = flag.Int("filesize", 1000, "Size of file in bytes")
	//binsize        = flag.Int("size of garbage bin", 10, "number of files in garbage bin") // TODO: Change to size of bin in bytes
)
const(
	binsize = 10
	filesize = "1000"
)

//########################################## The Actual KV store at this server ############################################################
type storeServer struct {
	kvs.UnimplementedStoreServer
	// key is 24 bytes, value is 10 bytes, but we are not enforcing it. 
	// Instaed we use a string for key, vlaue (note: '[]byte' can't be used as map key in golang)
	// Value is the path to a file
	store map[string]string 

	// All read, write require locking of entire KV store. BAD!
	lock         sync.Mutex 

	binLock     sync.Mutex 
	garbage      [binsize]string  //we don't immediately delete since a file might be being read by a client
	curBinsize  int
}

//####################################### gRPC Server Service implemmentation ###############################################################
// The Services are specified in kvstore/kvstore_file.proto

///************-------------------- GET Request Implementation -----------------------------**********************

// Source: https://stackoverflow.com/questions/58566016/how-can-i-serve-files-while-using-grpc

func (s *storeServer) GetTest(ctx context.Context, record *kvs.Record) (*kvs.Response, error) {

	//fmt.Println("Request rcvd")
	return &kvs.Response{FileChunk:"Hit",}, nil


}


func (s *storeServer) Get(record *kvs.Record, responseStream kvs.Store_GetServer) error {
    
	//bufferSize := 64 *1024 //64KiB, tweak this as desired
	bufferSize := 200 //64KiB, tweak this as desired
	s.lock.Lock()
		filename, ok := s.store[record.GetKey()]
		if ok {
			file, err := os.Open(filename) 
			if err != nil {
				fmt.Println(err)
				return err
			}
			defer file.Close()

			s.lock.Unlock()
			buff := make([]byte, bufferSize)
			for {
				bytesRead, err := file.Read(buff)
				if err != nil {
					if err != io.EOF {
						fmt.Println(err)
					}
					break
				}
				resp := &kvs.Response{
					FileChunk: string(buff[:bytesRead]),
				}
				err = responseStream.Send(resp)
				//log.Println(resp.GetFileChunk())
				if err != nil {
					log.Println("error while sending chunk:", err)
					return err
				}
			}
	} else {
		s.lock.Unlock() }
    return nil
}


///************-------------------- SET Request Implementation -----------------------------************************
// Request: SET (key, propoosed_value, proposed_ts). If key doesn't exist, add it. In any case, Ack to the sender.
func (s *storeServer) Set(ctx context.Context, record *kvs.Record) (*kvs.AckMsg, error) {
	//log.Printf("!~! Got a Set request for: %s\n", record.String())

	affected_key  := record.GetKey()
	currentTime := time.Now().UnixMilli()
	proposed_filename := fmt.Sprintf("dataset_files/files/file-%s-%d.dat",affected_key,currentTime)

	s.lock.Lock()
		current_filename, ok := s.store[affected_key]
		s.store[affected_key] = proposed_filename
		log.Printf(proposed_filename)

		_, err := exec.Command("bash", "generate_file.sh", proposed_filename, filesize ).Output()
		if err != nil {
			fmt.Printf("error %s", err)
		}

		if (ok && (current_filename!=proposed_filename)){ // if key was already present, then send the corresponding file to garbage
			s.binLock.Lock()
			s.curBinsize = (s.curBinsize + 1)%binsize
			cur_binfile := s.garbage[s.curBinsize]
			s.garbage[s.curBinsize] = current_filename
			
			if  cur_binfile!= "" { // bin is full
				log.Printf("Garbage bin is full, deleting the file %s", cur_binfile)
				e := os.Remove(cur_binfile)
				if e != nil {
					log.Fatal(e)
				}
			}
			s.binLock.Unlock()
		} else {
			// TODO if current_filename=proposed_filename, currently we are just replacing it immediately
		}
		// TODO: a separate garbage collector

	s.lock.Unlock()

	// Server Acks in any case.
	return  &kvs.AckMsg{}, nil
}

//TODO: Garbage collector

//#################################################### Utilities #############################################################################

// loadSome data into the KVstore to begin with
func (s *storeServer) loadKV() {
	
	for i := 1000; i < 1010; i++ {
		currentTime := time.Now().UnixMilli()
		fn := fmt.Sprintf("dataset_files/files/file-KEY%d-%d.dat",i,currentTime)
		_, err := exec.Command("bash", "generate_file.sh", fn, filesize ).Output()
		if err != nil {
			fmt.Printf("error %s \n", err)
		}
		keynew := fmt.Sprintf("%s%d", "KEY", i)
		kv1 := kvs.Record{
			Key: keynew ,
		}
		log.Println(kv1.GetKey())
		log.Println(fn)
		s.store[kv1.GetKey()] = fn
	}	
	log.Printf("Setup Complete.\n\n")
}


func newServer() *storeServer {
	s := &storeServer{ store: make(map[string]string),  curBinsize:0 }
	s.loadKV()
	return s
}


//################################################ MAIN ##################################################################


func main() {
	flag.Parse()
	log.Printf("Starting. Logging Type %s, Port no %s ", *logType, *port ) 
	//******************************************** SETUP LOGGING **************************************************
	
	switch *logType {
	case "file":
			fmt.Sprintf("Logging type: %s, find logs at logs/server-%s.txt \n", *logType, *port) 
			file, err := os.OpenFile( ("./logs/server-" + *port + ".txt"), os.O_CREATE|os.O_WRONLY, 0666 )
			if err != nil { log.Printf("Failed to open log: %s ", err) }
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