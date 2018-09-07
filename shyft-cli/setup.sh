#!/bin/bash
sh ./shyft-cli/resetShyftGeth.sh &&                     # Reset geth data - Remove pg and chain data
sh ./shyft-cli/initShyftGeth.sh                         # Init Shyft Geth
