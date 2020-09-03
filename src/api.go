package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jinzhu/copier"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/prometheus/common/log"
)

// Run : Start the node api server
func Run(port string) {
	api := &api{
		blockChain:     NewBlockChain(),
		nodeIdentifier: getUUID(),
		port:           port,
	}

	http.HandleFunc("/mine", api.mine)
	http.HandleFunc("/chain", api.chain)
	http.HandleFunc("/transactions/new", api.transactionsNew)
	http.HandleFunc("/getNodeUUID", api.getNodeUUID)
	http.HandleFunc("/nodes/register", api.registerNode)
	http.HandleFunc("/nodes/resolve", api.resolveChain)
	log.Info(fmt.Sprintf("Node UUID: %s", api.nodeIdentifier))
	log.Info(fmt.Sprintf("Listening on port %s", port))
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

// api : api for mutating the Blockchain
type api struct {
	blockChain     *BlockChain
	nodeIdentifier string
	port           string
}

func (a *api) getNodeUUID(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(a.nodeIdentifier))
}

func (a *api) mine(w http.ResponseWriter, req *http.Request) {
	lastBlock := a.blockChain.lastBlock()
	lastProof := lastBlock.Proof
	proof := a.blockChain.proofOfWork(lastProof)

	a.blockChain.newTransaction("0", a.nodeIdentifier, 1)
	previousHash := a.blockChain.hash(lastBlock)
	block := a.blockChain.newBlock(proof, previousHash)

	js, err := json.Marshal(block)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Trigger the other nodes to resolve the longest chain
	client := &http.Client{Timeout: 10 * time.Second}
	for _, node := range a.blockChain.nodes {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/nodes/resolve", node), nil)
		req.Header.Set("node-uuid", a.nodeIdentifier)
		_, err := client.Do(req)
		if err != nil {
			log.Error(fmt.Sprintf("Request to resolve chain on node %s failed, error: %s", node, err))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a *api) chain(w http.ResponseWriter, req *http.Request) {
	response := &ChainDTO{
		Chain:  a.blockChain.chain,
		Length: len(a.blockChain.chain),
	}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a *api) transactionsNew(w http.ResponseWriter, r *http.Request) {
	var transaction Transaction
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	a.blockChain.newTransaction(transaction.Sender, transaction.Recipient, transaction.Amount)

	// If request wasn't sent from another node, register list with all other nodes
	propagate := r.Header.Get("node-uuid") == ""
	if propagate {
		client := &http.Client{Timeout: 10 * time.Second}
		for _, node := range a.blockChain.nodes {
			messageBody, _ := json.Marshal(&transaction)
			req, _ := http.NewRequest("POST", fmt.Sprintf("%s/transactions/new", node), bytes.NewBuffer(messageBody))
			req.Header.Set("node-uuid", a.nodeIdentifier)

			_, err := client.Do(req)
			if err != nil {
				log.Error(fmt.Sprintf("Request to register nodes on node %s failed, error: %s", node, err))
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"message\":\"transaction added\"}"))
}

func (a *api) registerNode(w http.ResponseWriter, r *http.Request) {

	var nodes NodeDTO
	err := json.NewDecoder(r.Body).Decode(&nodes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	for _, node := range nodes.Nodes {
		parsedURL, _ := url.Parse(node)
		if parsedURL.Port() == a.port { // We don't want to register a node to itself, skip if the port is the same
			continue
		}
		a.blockChain.registerNode(node)
	}

	// If request wasn't sent from another node, register list with all other nodes
	propagate := r.Header.Get("node-uuid") == ""
	if propagate {
		client := &http.Client{Timeout: 10 * time.Second}
		for _, node := range a.blockChain.nodes {
			var blockChainNodes []string
			copier.Copy(&blockChainNodes, &a.blockChain.nodes) // deep copy

			// We need to include the current node for the request that will be sent to the newly registered node,
			// so it can register this node as well
			blockChainNodes = append(blockChainNodes, fmt.Sprintf("http://localhost:%s", a.port))

			messageBody, _ := json.Marshal(&NodeDTO{blockChainNodes})
			req, _ := http.NewRequest("POST", fmt.Sprintf("%s/nodes/register", node), bytes.NewBuffer(messageBody))
			req.Header.Set("node-uuid", a.nodeIdentifier)
			_, err := client.Do(req)
			if err != nil {
				log.Error(fmt.Sprintf("Request to register nodes on node %s failed, error: %s", node, err))
			}

		}
	}

	if len(nodes.Nodes) > 0 {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write([]byte("{\"message\":\"success\"}"))
}

func (a *api) resolveChain(w http.ResponseWriter, r *http.Request) {
	replaced := a.blockChain.resolveConfilcts()
	if replaced {
		log.Info(fmt.Sprintf("Chain replaced on node: %s", a.nodeIdentifier))
	} else {
		log.Info(fmt.Sprintf("Chain not replaced: %s", a.nodeIdentifier))

	}
	w.WriteHeader(http.StatusOK)
}

func getUUID() string {
	u, err := uuid.NewV4()
	if err != nil {
		log.Error("Unable to generate node uuid")
	}
	return u.String()
}
