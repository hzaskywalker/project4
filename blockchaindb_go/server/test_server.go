package main

import (
    "fmt"
    "math/rand"
    pb "../protobuf/go"
)

type FakeServer struct{
    people_id []string
    balance map[string]int
    channel chan *Transaction
}

func (s *FakeServer)init(n int, initilize int){
    s.balance = make(map[string]int)
    for i:=0;i<n;i++{
        s.people_id = append(s.people_id, fmt.Sprintf("%08x", i))
    }
    for i:=0;i<n;i++{
        s.balance[s.people_id[i]] = initilize
    }
}

func (s *FakeServer)GenTransfers(){
    n := len(s.people_id)
    for i:=0;i<10;i++{
        t:=NewTransaction()
        t.trans = &pb.Transaction{}
        t.trans.FromID = s.people_id[i%n]
        t.trans.ToID = s.people_id[(i+1)%n]
        t.trans.Value = int32(rand.Int() % 10)
        t.trans.MiningFee = 0
        t.trans.Type = pb.Transaction_Types(5)
        t.trans.UUID = fmt.Sprintf("%08x", i*234)
        t.UUID = t.trans.UUID

        s.channel <- t
    }
}

func (s *FakeServer)TRANSFER()*Transaction{
    //just test the producer and receiver
    //may send the same transfer several times
    //may ask something about transfer
    return <- s.channel
}

func testProducer(){
    s := FakeServer{}
    s.init(10, 100)
    go s.GenTransfers()
    fmt.Println(s.TRANSFER().UUID)
}

func TestServer() {
    fmt.Print("begin test server\n")
    testProducer()
}