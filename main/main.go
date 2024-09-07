package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"simple_blockchain/blockchain"
)

var blockChain *blockchain.BlockChain

func mineHandler(w http.ResponseWriter, r *http.Request) {

	if blockChain.GetCurrentTransactionsSize() == 0 {
		w.Write([]byte("current transaction empty"))
		return
	}

	//get proof of work
	proof := blockChain.GetPOW()

	//update transaction

	newTransaction := blockchain.Transaction{
		Sender:    "0",
		Recipient: blockChain.GetNodeIdentifier(),
		Amount:    1,
	}

	blockChain.NewTransaction(newTransaction)

	//new a block
	block := blockChain.NewBlock(proof)

	//return the new Block

	resp, _ := json.Marshal(block)
	w.Write(resp)
	return
}

type FullChainResp struct {
	Chain        []blockchain.Block       `json:"chain"`
	Len          int                      `json:"len"`
	Transactions []blockchain.Transaction `json:"transactions"`
	Nodes        []string                 `json:"nodes"`
}

func fullChainHandler(w http.ResponseWriter, r *http.Request) {
	resp := &FullChainResp{
		Chain:        blockChain.Chain,
		Len:          len(blockChain.Chain),
		Transactions: blockChain.CurrentTransactions,
		Nodes:        blockChain.Nodes,
	}

	fmt.Println("full chain handler : ", blockChain.Chain)

	respBytes, _ := json.Marshal(resp)
	w.Write(respBytes)
	return
}

// new a transacation into block chain, need to be mind
func newTransactionHandler(w http.ResponseWriter, r *http.Request) {
	//parse body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("fail to read request body")
		w.Write([]byte("Fail to read request body"))
		return
	}

	var req blockchain.Transaction
	err = json.Unmarshal(bodyBytes, &req)

	if err != nil {
		fmt.Println("Fail to convert request body into transaction Req")
		fmt.Println("body = ", string(bodyBytes))
		w.Write([]byte("Fail to convert request body into transaction Req"))
		return
	}

	blockChain.NewTransaction(req)
	w.Write([]byte(fmt.Sprintf("the transaction will be added to index = %d block", len(blockChain.Chain))))
	return
}

type RegisterNodeReq struct {
	Node string `json:"node"`
}

// register a neighbour node into the block chain
func registerNodeHandler(w http.ResponseWriter, r *http.Request) {

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Fail to read request body")
		w.Write([]byte("Fail to read request body"))
		return
	}

	var req RegisterNodeReq
	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		fmt.Println("Fail to convert request body into transaction Req")
		fmt.Println("body = ", string(bodyBytes))
		w.Write([]byte("Fail to convert request body into transaction Req"))
		return
	}

	blockChain.RegisterNode(req.Node)
	nodes := blockChain.GetNodes()
	nodesStr := ""
	for i := 0; i < len(nodes); i++ {
		nodesStr = nodesStr + nodes[i] + " "
	}

	w.Write([]byte(nodesStr))
	return
}

// resloeve conflicts with neighbour nodes
func resolveConflictsHandler(w http.ResponseWriter, r *http.Request) {
	blockChain.ResolveConflicts()
	chainBytes, _ := json.Marshal(blockChain.Chain)
	w.Write(chainBytes)
	return
}

func main() {
	var nodeIdentifier string
	var port string
	flag.StringVar(&nodeIdentifier, "ni", "", "node identifier")
	flag.StringVar(&port, "p", "", "http port")

	flag.Parse()
	fmt.Println("nodeIdentifier = ", nodeIdentifier)
	fmt.Println("port = ", port)

	blockChain = &blockchain.BlockChain{}
	blockChain.Init(nodeIdentifier)

	http.HandleFunc("/fullChain", fullChainHandler)
	http.HandleFunc("/newTransaction", newTransactionHandler)
	http.HandleFunc("/mine", mineHandler)
	http.HandleFunc("/registerNode", registerNodeHandler)
	http.HandleFunc("/resolveConflicts", resolveConflictsHandler)

	go func() {
		http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	}()

	select {}

}
