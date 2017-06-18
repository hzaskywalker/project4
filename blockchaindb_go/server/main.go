package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net"
    "flag"
    
    pb "../protobuf/go"
    "../hash"

    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

const maxBlockSize = 50

type server struct{
	s *Service
}

// Database Interface 
func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
    //return &pb.GetResponse{Value: 1000}, nil
	return s.s.Get(in)
}
func (s *server) Transfer(ctx context.Context, in *pb.Transaction) (*pb.BooleanResponse, error) {
    //return &pb.BooleanResponse{Success: true}, nil
	return s.s.Transfer(in)
}
func (s *server) Verify(ctx context.Context, in *pb.Transaction) (*pb.VerifyResponse, error) {
    //return &pb.VerifyResponse{Result: pb.VerifyResponse_FAILED, BlockHash:"?"}, nil
	return s.s.Verify(in)
}
func (s *server) PushTransaction(ctx context.Context, in *pb.Transaction) (*pb.Null, error) {
    //return &pb.Null{}, nil
	return s.s.PushTransaction(in)
}
func (s *server) PushBlock(ctx context.Context, in *pb.JsonBlockString) (*pb.Null, error) {
    //return &pb.Null{}, nil
	return s.s.PushBlock(in)
}
func (s *server) GetHeight(ctx context.Context, in *pb.Null) (*pb.GetHeightResponse, error) {
    return s.s.GetHeight(in)
}
func (s *server) GetBlock(ctx context.Context, in *pb.GetBlockRequest) (*pb.JsonBlockString, error) {
    return s.s.GetBlock(in)
}



var id=flag.Int("id",1,"Server's ID, 1<=ID<=NServers")
var Dat map[string]interface{}
var IDstr string
// Main function, RPC server initialization
func main() {
    //set the hardness
    HashHardness = 5
    InitHash = "0000000000000000000000000000000000000000000000000000000000000000"

    //TestDatabase()
    //TestMiner()
    TestMainLoop()
    return

    flag.Parse()
    IDstr = fmt.Sprintf("%d",*id)

    _=fmt.Sprintf("Server%02d",*id)
    _=hash.GetHashString
    

    // Read config
	conf, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(conf, &Dat)
	if err != nil {
		panic(err)
	}
	dat := Dat[IDstr].(map[string]interface{}) // should be dat[myNum] in the future
	address, _ := fmt.Sprintf("%s:%s", dat["ip"], dat["port"]), fmt.Sprintf("%s",dat["dataDir"])
	

    // Bind to port
    lis, err := net.Listen("tcp", address)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    log.Printf("Listening: %s ...", address)

    // Create gRPC server
	grpc_s := grpc.NewServer()
	q := RealServer{rpc:grpc_s}
	miner := NewMiner(&q)

    s := &server{}
	s.s = NewService()
	s.s.transfer = miner.transfers
    pb.RegisterBlockChainMinerServer(grpc_s, s)
    // Register reflection service on gRPC server.
    reflection.Register(grpc_s)
	go miner.mainLoop(s.s)

    // Start server
    if err := grpc_s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}



