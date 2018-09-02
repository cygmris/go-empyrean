package main

///@NOTE Shyft handler functions when endpoints are hit
import (
	"fmt"
	"net/http"

	_ "github.com/lib/pq"

	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/ShyftNetwork/go-empyrean/core"
	"github.com/gorilla/mux"
)

// GetTransaction gets txs
func GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txHash := vars["txHash"]
	sqldb, err := core.DBConnection()
	getTxResponse := core.SGetTransaction(sqldb, txHash)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, getTxResponse)
}

// GetAllTransactions gets txs
func GetAllTransactions(w http.ResponseWriter, r *http.Request) {

	sqldb, err := core.DBConnection()

	txs := core.SGetAllTransactions(sqldb)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, txs)
}

// GetAllTransactions gets txs
func GetAllTransactionsFromBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockNumber := vars["blockNumber"]

	sqldb, err := core.DBConnection()

	txsFromBlock := core.SGetAllTransactionsFromBlock(sqldb, blockNumber)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, txsFromBlock)
}

func GetAllBlocksMinedByAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	coinbase := vars["coinbase"]

	sqldb, err := core.DBConnection()

	blocksMined := core.SGetAllBlocksMinedByAddress(sqldb, coinbase)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, blocksMined)
}

// GetAccount gets balance
func GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	sqldb, err := core.DBConnection()

	getAccountBalance := core.SGetAccount(sqldb, address)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, getAccountBalance)
}

// GetAccount gets balance
func GetAccountTxs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	sqldb, err := core.DBConnection()

	getAccountTxs := core.SGetAccountTxs(sqldb, address)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, getAccountTxs)
}

// GetAllAccounts gets balances
func GetAllAccounts(w http.ResponseWriter, r *http.Request) {

	sqldb, err := core.DBConnection()

	allAccounts := core.SGetAllAccounts(sqldb)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, allAccounts)
}

//GetBlock returns block json
func GetBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockNumber := vars["blockNumber"]

	sqldb, err := core.DBConnection()
	getBlockResponse := core.SGetBlock(sqldb, blockNumber)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, getBlockResponse)
}

// GetAllBlocks response
func GetAllBlocks(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request -> ", r)
	sqldb, err := core.DBConnection()
	block3 := core.SGetAllBlocks(sqldb)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, block3)
}

func GetRecentBlock(w http.ResponseWriter, r *http.Request) {

	sqldb, err := core.DBConnection()

	mostRecentBlock := core.SGetRecentBlock(sqldb)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, mostRecentBlock)

}

//GetInternalTransactions gets internal txs
func GetInternalTransactionsByHash(w http.ResponseWriter, r *http.Request) {
	sqldb, err := core.DBConnection()

	vars := mux.Vars(r)
	txHash := vars["txHash"]

	internalTxs := core.SGetInternalTransaction(sqldb, txHash)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, internalTxs)
}

//GetInternalTransactionsHash gets internal txs hash
func GetInternalTransactions(w http.ResponseWriter, r *http.Request) {
	sqldb, err := core.DBConnection()

	internalTxs := core.SGetAllInternalTransactions(sqldb)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, internalTxs)
}

func BroadcastTx(w http.ResponseWriter, r *http.Request) {
	// Example return result (returns tx hash):
	// {"jsonrpc":"2.0","id":1,"result":"0xafa4c62f29dbf16bbfac4eea7cbd001a9aa95c59974043a17f863172f8208029"}

	// http params
	vars := mux.Vars(r)
	transactionHash := vars["transaction_hash"]

	// format the transactionHash into a proper sendRawTransaction jsonrpc request
	formatted_json := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["%s"],"id":0}`, transactionHash))

	// send json rpc request
	resp, _ := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(formatted_json))
	body, _ := ioutil.ReadAll(resp.Body)
	byt := []byte(string(body))

	// read json and return result as http response, be it an error or tx hash
	var dat map[string]interface{}
	if err := json.Unmarshal(byt, &dat); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ERROR parsing json")
	}
	tx_hash := dat["result"]
	if tx_hash == nil {
		errMap := dat["error"].(map[string]interface{})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ERROR:", errMap["message"])
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Transaction Hash:", tx_hash)
	}
}

func GetTXPoolContent() interface{} {
	jsonStr := fmt.Sprintf(`{"jsonrpc":"2.0","method":"txpool_content","params":[],"id":67}`)
	jsonBytes := []byte(jsonStr)
	fmt.Println(string(jsonBytes))

	req, err := http.NewRequest("POST", "http://localhost:8545", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fmt.Println("Error: ", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Println("body")
	//fmt.Println(resp.Body)

	var target map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&target)
	fmt.Println(target["result"])

	return target["result"]

	///fmt.Println(target["result"])
}

//GetInternalTransactionsHash gets internal txs hash
func WriteTXPool(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HERE")

	foo := GetTXPoolContent()
	fmt.Println("foo ", foo)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	//fmt.Fprintln(w, "{\"foo\": 12}")
	jsonString, err := json.Marshal(foo)
	if err != nil {
		fmt.Println("err is ", err)
	}
	fmt.Fprintln(w, string(jsonString[:]))

}
