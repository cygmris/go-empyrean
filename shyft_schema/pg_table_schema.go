package shyftschema

import (
	"fmt"
)

// This Package Contains The Queries To Make Tables Required For the PG Database

// TableQuery - returns sql to create db tables
func MakeTableQuery() string {
	return fmt.Sprintf(`%s %s %s %s`, blocksTable, txsTable, accountsTable, internalTxsTable)
}

// AccountsTable sql for accounts
const accountsTable = `
CREATE TABLE IF NOT EXISTS accounts (
  addr text primary key unique,
  balance numeric,
  accountNonce numeric
);
`

// BlocksTable sql for Blocks
const blocksTable = `
CREATE TABLE IF NOT EXISTS blocks (
  hash text primary key,
  coinbase text,
  gasUsed numeric,
  gasLimit numeric,
  txCount numeric,
  uncleCount numeric,
  age timestamp,
  parentHash text,
  uncleHash text,
  difficulty bigint,
  size text,
  nonce numeric,
  rewards numeric,
  number bigint
);
`

// TxsTable sql for transactions
const txsTable = `
CREATE TABLE IF NOT EXISTS txs (
  txHash text primary key unique,
  to_addr text,
  from_addr text,
  blockhash text references blocks(hash),
  blocknumber text,
  amount numeric,
  gasprice numeric,
  gas numeric,
  gasLimit numeric,
  txFee numeric,
  nonce numeric,
  txStatus text,
  isContract bool,
  age timestamp,
  data bytea
);
`

// InternalTxsTable sql for transactions
const internalTxsTable = `
CREATE TABLE IF NOT EXISTS internalTxs (
  id SERIAL PRIMARY KEY,
  txHash text references txs(txHash),
  action text,
  to_addr text,
  from_addr text,
  amount text,
  gas numeric,
  gasUsed numeric,
  time text,
  input text,
  output text
);
`
