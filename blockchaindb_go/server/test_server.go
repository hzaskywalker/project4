package main

import (
    "fmt"
    "math/rand"
    "time"
    "sort"
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

func (s *FakeServer)GenTransfers(start int){
    n := len(s.people_id)
    for i:=0;i<50;i++{
        t:=NewTransaction()
        t.trans = &pb.Transaction{}
        t.trans.FromID = s.people_id[i%n]
        t.trans.ToID = s.people_id[(i+1)%n]
        t.trans.Value = int32(rand.Int() % 10)
        t.trans.Value = int32(i)
        t.trans.MiningFee = 0
        t.trans.Type = pb.Transaction_Types(5)
        t.trans.UUID = fmt.Sprintf("%08x", i + start*234)
        t.UUID = t.trans.UUID

        s.channel <- t
    }
}

func (s *FakeServer)TRANSFER()*Transaction{
    //just test the producer and receiver
    //may send the same transfer several times
    //may ask something about transfer
    t := <- s.channel
    return t
}

func testProducer(){
    s := FakeServer{channel:make(chan *Transaction)}

    //check send block and receive block

    fakeProducer := func (T* TransferManager){
        for i:=0;i<50;i++ {
            t := T.server.TRANSFER()
            T.ReadWriteTransaction(t)
        }
    }

    s.init(10, 100)
    go s.GenTransfers(0)

    transferManager := NewTransferManager(&s)
    go fakeProducer(transferManager)

    block := transferManager.GetBlockSync()
    fmt.Println("Block hash: ", block.GetHash())
    fmt.Println("ramaining Pending size ", len(transferManager.Pending))

    fmt.Println("Go GenTransfers again")

    go s.GenTransfers(0)
    fakeProducer(transferManager)
    fmt.Println("after producer2", len(transferManager.Pending))

    transferManager.UpdateBlockStatus(block, 3)
    s.channel = make(chan *Transaction, 50)
    go s.GenTransfers(0)
    fakeProducer(transferManager)

    block = transferManager.GetBlockSync()

    keys := []string{}
    dict := make(map[string]*pb.Transaction)
    for _, tt := range block.Transactions {
        keys = append(keys, tt.UUID)
        dict[tt.UUID] = tt
    }
    sort.Strings(keys)
    for i:=0;i<50;i++{
        block.Transactions[i] = dict[keys[i]]
    }
    fmt.Println("Block hash(This should be the same for each run):\n", block.GetHash())

    transferManager.UpdateBlockStatus(block, 3)
    go s.GenTransfers(0)
    go s.GenTransfers(0)
    go s.GenTransfers(0)
    go fakeProducer(transferManager)
    go fakeProducer(transferManager)
    go fakeProducer(transferManager)
    go fakeProducer(transferManager)
    go fakeProducer(transferManager)
    go fakeProducer(transferManager)

    receiver := func (){
        block = transferManager.GetBlockSync()
        fmt.Println("Muliple producer Hash:", block.GetHash())
    }
    receiver()
    go receiver()
    fmt.Println("time.Second")
    _ = <- time.After(time.Second)
    fmt.Println("Dict size", transferManager.GetDictSize())
    go s.GenTransfers(1)
    _ = <- time.After(time.Second)
}

func TestServer() {
    fmt.Print("begin test server\n")
    testProducer()
}