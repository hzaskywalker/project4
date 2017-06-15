/*
    network server
    and in fact cache the network results into disk for fast recover
    and check the some validity: types check
    */
package main

type Server struct{
}

func NewServer()*Server{
    return &Server{}
}

func (s *Server)GetHeight()(int, *Block, bool){
    return 0, &Block{}, false
}

func (s *Server)GetBlock(hash string)(*Block, bool){
    return &Block{}, false
}
func (s *Server)TRANSFER()*Transaction{
    return &Transaction{}
}