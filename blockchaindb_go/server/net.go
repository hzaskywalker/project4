/*
    network server
    and in fact cache the network results into disk for fast recover
    and check the some validity: types check
    */
package main

import (
    pb "../protobuf/go"
)


type Service struct{
    //To receive other's quest
    GetRequest chan string
    GetResponse chan int
    //TRANSFER //handeled by transfer
    VerifyRequest chan string
    VerifyResponse chan *Block
    //PushTransaction 
    GetHeightRequest chan bool
    GetHeightResponse chan *Block

    GetBlockRequest chan string
    GetBlockReponse chan *Block

    PushBlockRequest chan *Block
    PushBlockResponse chan bool
}

func NewService() *Service{
    s := &Service{}
    s.GetRequest = make(chan string)
    s.GetResponse = make(chan int)
    s.VerifyRequest = make(chan string)
    s.VerifyResponse = make(chan *Block)

    s.GetHeightRequest = make(chan bool)
    s.GetHeightResponse = make(chan *Block)

    s.GetBlockRequest = make(chan string)
    s.GetBlockReponse = make(chan *Block)

    s.PushBlockRequest = make(chan *Block)
    s.PushBlockResponse = make(chan bool)
    return s
}

func (s *Service) Get(q *pb.GetRequest)*pb.GetResponse{
    s.GetRequest <- q.UserID
    return &pb.GetResponse{Value: int32(<-s.GetResponse)}
}

/*
func (s *Service) Transfer(in *pb.Transaction) (*pb.BooleanResponse, error) {
    return &pb.BooleanResponse{Success: true}, nil
}
*/

func (s *Service) Verify(in *pb.Transaction) (*pb.VerifyResponse, error) {
    //We don't need to check other things
    s.VerifyRequest <- in.UUID
    ok := <- s.VerifyResponse
    return &pb.VerifyResponse{Result: pb.VerifyResponse_FAILED, BlockHash:ok.GetHash()}, nil
}

func (s *Service) PushBlock(in *pb.JsonBlockString) (*pb.Null, error) {
    block := MakeNewBlock()
    block.Unmarshal(in.Json)

    s.PushBlockRequest <- block
    <- s.PushBlockResponse
    //need broad cast
    return &pb.Null{}, nil
}

func (s *Service) GetBlock(in *pb.GetBlockRequest) (*pb.JsonBlockString, error) {
    s.GetBlockRequest <- in.BlockHash
    block := <- s.GetBlockReponse
    if block == nil{
        return nil, nil
    } else{
        return &pb.JsonBlockString{Json:block.MarshalToString()}, nil
    }
}

type Server interface{
    GetHeight()(int, *Block, bool)
    GetBlock(hash string)(*Block, bool)
    TRANSFER()*Transaction

    GetBlocksByBalance(*DatabaseEngine, chan *Block, chan int)
}

type RealServer struct{
}
func (*RealServer)GetHeight()(int, *Block, bool){return 0, &Block{}, false}
func (*RealServer)GetBlock(hash string)(*Block, bool){return &Block{}, false}
func (*RealServer)TRANSFER()*Transaction{return &Transaction{}}
