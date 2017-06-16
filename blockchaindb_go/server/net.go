/*
    network server
    and in fact cache the network results into disk for fast recover
    and check the some validity: types check
    */
package main

type Server interface{
    GetHeight()(int, *Block, bool)
    GetBlock(hash string)(*Block, bool)
    TRANSFER()*Transaction
}

type RealServer struct{
}
func (*RealServer)GetHeight()(int, *Block, bool){return 0, &Block{}, false}
func (*RealServer)GetBlock(hash string)(*Block, bool){return &Block{}, false}
func (*RealServer)TRANSFER()*Transaction{return &Transaction{}}
