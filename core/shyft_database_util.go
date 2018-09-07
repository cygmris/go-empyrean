package core

import (
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ShyftNetwork/go-empyrean/common"
	Rewards "github.com/ShyftNetwork/go-empyrean/consensus/ethash"
	stypes "github.com/ShyftNetwork/go-empyrean/core/sTypes"
	"github.com/ShyftNetwork/go-empyrean/core/types"
	"github.com/ShyftNetwork/go-empyrean/shyfttracerinterface"
	_ "github.com/lib/pq"
)

//IShyftTracer Used to initialize ShyftTracer
var IShyftTracer shyfttracerinterface.IShyftTracer

//SetIShyftTracer sets tracer type
func SetIShyftTracer(st shyfttracerinterface.IShyftTracer) {
	IShyftTracer = st
}

//SWriteBlock writes to block info to sql db
func SWriteBlock(block *types.Block, receipts []*types.Receipt) error {
	//Get miner rewards
	rewards := swriteMinerRewards(block)
	//Format block time to be stored
	i, err := strconv.ParseInt(block.Time().String(), 10, 64)
	if err != nil {
		panic(err)
	}
	age := time.Unix(i, 0)

	blockData := stypes.SBlock{
		Hash:       block.Header().Hash().Hex(),
		Coinbase:   block.Header().Coinbase.String(),
		Number:     block.Header().Number.String(),
		GasUsed:    block.Header().GasUsed,
		GasLimit:   block.Header().GasLimit,
		TxCount:    block.Transactions().Len(),
		UncleCount: len(block.Uncles()),
		ParentHash: block.ParentHash().String(),
		UncleHash:  block.UncleHash().String(),
		Difficulty: block.Difficulty().String(),
		Size:       block.Size().String(),
		Nonce:      block.Nonce(),
		Rewards:    rewards,
		Age:        age,
	}

	//Inserts block data into DB
	InsertBlock(blockData)

	if block.Transactions().Len() > 0 {
		for _, tx := range block.Transactions() {
			swriteTransactions(tx, block.Header().Hash(), blockData.Number, receipts, age, blockData.GasLimit)
		}
	}
	return nil
}

//swriteTransactions writes to sqldb, a SHYFT postgres instance
func swriteTransactions(tx *types.Transaction, blockHash common.Hash, blockNumber string, receipts []*types.Receipt, age time.Time, gasLimit uint64) error {
	var isContract bool
	var statusFromReciept, toAddr string
	var contractAddressFromReciept common.Address
	if tx.To() == nil {
		for _, receipt := range receipts {
			statusReciept := (*types.ReceiptForStorage)(receipt).Status
			contractAddressFromReciept = (*types.ReceiptForStorage)(receipt).ContractAddress
			switch {
			case statusReciept == 0:
				statusFromReciept = "FAIL"
			case statusReciept == 1:
				statusFromReciept = "SUCCESS"
			}
		}
		isContract = true
		tempAddr := &contractAddressFromReciept
		toAddr = tempAddr.String()
	} else {
		isContract = false
		for _, receipt := range receipts {
			statusReciept := (*types.ReceiptForStorage)(receipt).Status
			switch {
			case statusReciept == 0:
				statusFromReciept = "FAIL"
			case statusReciept == 1:
				statusFromReciept = "SUCCESS"
			}
		}
		toAddr = tx.To().String()
	}

	txData := stypes.ShyftTxEntryPretty{
		TxHash:      tx.Hash().Hex(),
		From:        tx.From().Hex(),
		To:          toAddr,
		BlockHash:   blockHash.Hex(),
		BlockNumber: blockNumber,
		Amount:      tx.Value().String(),
		Cost:        tx.Cost().Uint64(),
		GasPrice:    tx.GasPrice().Uint64(),
		GasLimit:    gasLimit,
		Gas:         tx.Gas(),
		Nonce:       tx.Nonce(),
		Age:         age,
		Data:        tx.Data(),
		Status:      statusFromReciept,
		IsContract:  isContract,
	}
	isContractCheck := IsContract(txData.To)
	if isContractCheck == true {
		InsertTx(txData)
		//Runs necessary functions for tracing internal transactions through tracers.go
		IShyftTracer.GetTracerToRun(tx.Hash())
	} else {
		//Inserts Tx into DB
		InsertTx(txData)
	}
	return nil
}

//SWriteInternalTxBalances Writes internal txs and updates balances
func SWriteInternalTxBalances(sqldb *sql.DB, toAddr string, fromAddr string, amount string) error {
	sendAndReceiveData := stypes.SendAndReceive{
		To:     toAddr,
		From:   fromAddr,
		Amount: amount,
	}
	_, _, err := AccountExists(sendAndReceiveData.To)
	value := new(big.Int)
	value, _ = value.SetString(amount, 10)
	switch {
	case err == sql.ErrNoRows:
		accountNonce := "1"
		CreateAccount(sendAndReceiveData.To, sendAndReceiveData.Amount, accountNonce)
		adjustBalanceFromAddr(sendAndReceiveData, value)
	case err != nil:
		log.Fatal(err)
	default:
		balanceHelper(sendAndReceiveData, amount)
	}
	return nil
}

func adjustBalanceFromAddr(s stypes.SendAndReceive, value *big.Int) {
	fromAddressBalance, fromAccountNonce, err := AccountExists(s.From)
	switch {
	case err == sql.ErrNoRows:
		CreateAccount(s.From, "0", "1")
		fmt.Println("New From account created")
	}
	if err != nil {
		log.Fatal(err)
	}
	var newBalanceSender, newAccountNonceSender big.Int
	var nonceIncrement = big.NewInt(1)

	fromBalance := new(big.Int)
	fromBalance, _ = fromBalance.SetString(fromAddressBalance, 10)

	fromNonce := new(big.Int)
	fromNonce, _ = fromNonce.SetString(fromAccountNonce, 10)

	newBalanceSender.Sub(fromBalance, value)
	newAccountNonceSender.Add(fromNonce, nonceIncrement)

	UpdateAccount(s.From, newBalanceSender.String(), newAccountNonceSender.String())
}

func balanceHelper(s stypes.SendAndReceive, amount string) {
	fromAddressBalance, fromAccountNonce, err := AccountExists(s.From)
	toAddressBalance, toAccountNonce, err := AccountExists(s.To)
	if err != nil {
		log.Fatal(err)
	}
	var newBalanceReceiver, newBalanceSender, newAccountNonceReceiver, newAccountNonceSender big.Int
	var nonceIncrement = big.NewInt(1)

	//STRING TO BIG INT
	//BALANCES TO AND FROM ADDR
	toBalance := new(big.Int)
	toBalance, _ = toBalance.SetString(toAddressBalance, 10)

	fromBalance := new(big.Int)
	fromBalance, _ = fromBalance.SetString(fromAddressBalance, 10)

	amountValue := new(big.Int)
	amountValue, _ = amountValue.SetString(amount, 10)

	//ACCOUNT NONCES
	toNonce := new(big.Int)
	toNonce, _ = toNonce.SetString(toAccountNonce, 10)

	fromNonce := new(big.Int)
	fromNonce, _ = fromNonce.SetString(fromAccountNonce, 10)

	newBalanceReceiver.Add(toBalance, amountValue)
	newBalanceSender.Sub(fromBalance, amountValue)

	newAccountNonceReceiver.Add(toNonce, nonceIncrement)
	newAccountNonceSender.Add(fromNonce, nonceIncrement)

	//UPDATE ACCOUNTS BASED ON NEW BALANCES AND ACCOUNT NONCES
	UpdateAccount(s.To, newBalanceReceiver.String(), newAccountNonceReceiver.String())
	UpdateAccount(s.From, newBalanceSender.String(), newAccountNonceSender.String())
}

// @NOTE: This function is extremely complex and requires heavy testing and knowdlege of edge cases:
// uncle blocks, account balance updates based on reorgs, diverges that get dropped.
// Reason for this is because the accounts are not deterministic like the block and tx hashes.
// @TODO: Calculate reorg
func swriteMinerRewards(block *types.Block) string {
	minerAddr := block.Coinbase().String()
	shyftConduitAddress := Rewards.ShyftNetworkConduitAddress.String()
	// Calculate the total gas used in the block
	totalGas := new(big.Int)
	for _, tx := range block.Transactions() {
		totalGas.Add(totalGas, new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(tx.Gas())))
	}

	totalMinerReward := totalGas.Add(totalGas, Rewards.ShyftMinerBlockReward)

	// References:
	// https://ethereum.stackexchange.com/questions/27172/different-uncles-reward
	// line 551 in consensus.go (go-empyrean/consensus/ethash/consensus.go)
	// Some weird constants to avoid constant memory allocs for them.
	var big8 = big.NewInt(8)
	var uncleRewards []*big.Int
	var uncleAddrs []string

	// uncleReward is overwritten after each iteration
	uncleReward := new(big.Int)
	for _, uncle := range block.Uncles() {
		uncleReward.Add(uncle.Number, big8)
		uncleReward.Sub(uncleReward, block.Number())
		uncleReward.Mul(uncleReward, Rewards.ShyftMinerBlockReward)
		uncleReward.Div(uncleReward, big8)
		uncleRewards = append(uncleRewards, uncleReward)
		uncleAddrs = append(uncleAddrs, uncle.Coinbase.String())
	}

	sstoreReward(minerAddr, totalMinerReward)
	sstoreReward(shyftConduitAddress, Rewards.ShyftNetworkBlockReward)
	var uncRewards = new(big.Int)
	for i := 0; i < len(uncleAddrs); i++ {
		_ = uncleRewards[i]
		sstoreReward(uncleAddrs[i], uncleRewards[i])
	}

	fullRewardValue := new(big.Int)
	fullRewardValue.Add(totalMinerReward, Rewards.ShyftNetworkBlockReward)
	fullRewardValue.Add(fullRewardValue, uncRewards)

	return fullRewardValue.String()
}

func sstoreReward(address string, reward *big.Int) {
	// Check if address exists
	addressBalance, accountNonce, err := AccountExists(address)

	if err == sql.ErrNoRows {
		// Addr does not exist, thus create new entry
		// We convert totalReward into a string and postgres converts into number
		CreateAccount(address, reward.String(), "1")
		return
	} else if err != nil {
		// Something went wrong panic
		panic(err)
	} else {
		// Addr exists, update existing balance
		bigBalance := new(big.Int)
		var nonceIncrement = big.NewInt(1)
		currentAccountNonce := new(big.Int)
		currentAccountNonce, errorr := currentAccountNonce.SetString(accountNonce, 10)
		if !errorr {
			panic(errorr)
		}
		bigBalance, err := bigBalance.SetString(addressBalance, 10)
		if !err {
			panic(err)
		}
		newBalance := new(big.Int)
		newAccountNonce := new(big.Int)
		newBalance.Add(newBalance, bigBalance)
		newBalance.Add(newBalance, reward)
		newAccountNonce.Add(currentAccountNonce, nonceIncrement)
		//Update the balance and nonce
		UpdateAccount(address, newBalance.String(), newAccountNonce.String())
		return
	}
}

///////////////////////
//DB Utility functions
//////////////////////

//CreateAccount writes new account to Postgres Db
func CreateAccount(addr string, balance string, accountNonce string) {
	sqldb, _ := DBConnection()
	sqlStatement := `INSERT INTO accounts(addr, balance, accountNonce) VALUES(($1), ($2), ($3)) RETURNING addr;`
	insertErr := sqldb.QueryRow(sqlStatement, strings.ToLower(addr), balance, accountNonce).Scan(&addr)
	if insertErr != nil {
		panic(insertErr)
	}
}

//AccountExists checks if account exists in Postgres Db
func AccountExists(addr string) (string, string, error) {
	sqldb, _ := DBConnection()
	var addressBalance, accountNonce string
	sqlExistsStatement := `SELECT balance, accountNonce from accounts WHERE addr = ($1);`
	err := sqldb.QueryRow(sqlExistsStatement, strings.ToLower(addr)).Scan(&addressBalance, &accountNonce)
	switch {
	case err == sql.ErrNoRows:
		return addressBalance, accountNonce, err
	case err != nil:
		panic(err)
	default:
		return addressBalance, accountNonce, err
	}
}

//BlockExists checks if block exists in Postgres Db
func BlockExists(hash string) bool {
	var res bool
	sqlExistsStatement := `SELECT exists(select hash from blocks WHERE hash= ($1));`
	sqldb, _ := DBConnection()
	err := sqldb.QueryRow(sqlExistsStatement, strings.ToLower(hash)).Scan(&res)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			panic(err)
		}
	}
	return res
}

//IsContract checks if toAddr is from a contract in Postgres Db
func IsContract(addr string) bool {
	sqldb, _ := DBConnection()
	var isContract bool
	sqlExistsStatement := `SELECT isContract from txs WHERE to_addr=($1)`
	err := sqldb.QueryRow(sqlExistsStatement, strings.ToLower(addr)).Scan(&isContract)
	switch {
	case err == sql.ErrNoRows:
		return isContract
	default:
		return isContract
	}
}

//UpdateAccount updates account in Postgres Db
func UpdateAccount(addr string, balance string, accountNonce string) {
	sqldb, _ := DBConnection()
	updateSQLStatement := `UPDATE accounts SET balance = ($2), accountNonce = ($3) WHERE addr = ($1)`
	_, updateErr := sqldb.Exec(updateSQLStatement, strings.ToLower(addr), balance, accountNonce)
	if updateErr != nil {
		panic(updateErr)
	}
}

//InsertBlock writes block to Postgres Db
func InsertBlock(blockData stypes.SBlock) {
	sqldb, _ := DBConnection()
	sqlStatement := `INSERT INTO blocks(hash, coinbase, number, gasUsed, gasLimit, txCount, uncleCount, age, parentHash, uncleHash, difficulty, size, rewards, nonce) VALUES(($1), ($2), ($3), ($4), ($5), ($6), ($7), ($8), ($9), ($10), ($11), ($12),($13), ($14)) RETURNING number`
	qerr := sqldb.QueryRow(sqlStatement, strings.ToLower(blockData.Hash), blockData.Coinbase, blockData.Number, blockData.GasUsed, blockData.GasLimit, blockData.TxCount, blockData.UncleCount, blockData.Age, blockData.ParentHash, blockData.UncleHash, blockData.Difficulty, blockData.Size, blockData.Rewards, blockData.Nonce).Scan(&blockData.Number)
	if qerr != nil {
		panic(qerr)
	}
}

//InsertTx writes tx to Postgres Db
func InsertTx(txData stypes.ShyftTxEntryPretty) {
	sqldb, _ := DBConnection()
	var retNonce string
	sqlStatement := `INSERT INTO txs(txhash, from_addr, to_addr, blockhash, blockNumber, amount, gasprice, gas, gasLimit, txfee, nonce, isContract, txStatus, age, data) VALUES(($1), ($2), ($3), ($4), ($5), ($6), ($7), ($8), ($9), ($10), ($11), ($12), ($13), ($14), ($15)) RETURNING nonce`
	err := sqldb.QueryRow(sqlStatement, strings.ToLower(txData.TxHash), strings.ToLower(txData.From), strings.ToLower(txData.To), strings.ToLower(txData.BlockHash), txData.BlockNumber, txData.Amount, txData.GasPrice, txData.Gas, txData.GasLimit, txData.Cost, txData.Nonce, txData.IsContract, txData.Status, txData.Age, txData.Data).Scan(&retNonce)
	if err != nil {
		panic(err)
	}
}

//InsertInternalTx writes internal tx to Postgres Db
func InsertInternalTx(sqldb *sql.DB, i stypes.InteralWrite) {
	var returnValue string
	sqlStatement := `INSERT INTO internaltxs(action, txhash, from_addr, to_addr, amount, gas, gasUsed, time, input, output) VALUES(($1), ($2), ($3), ($4), ($5), ($6), ($7), ($8), ($9), ($10)) RETURNING txHash`
	qerr := sqldb.QueryRow(sqlStatement, i.Action, strings.ToLower(i.Hash), strings.ToLower(i.From), strings.ToLower(i.To), i.Value, i.Gas, i.GasUsed, i.Time, i.Input, i.Output).Scan(&returnValue)
	if qerr != nil {
		fmt.Println(qerr)
		panic(qerr)
	}
}
