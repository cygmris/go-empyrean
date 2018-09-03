// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package bind_test

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ShyftNetwork/go-empyrean/accounts/abi/bind"
	"github.com/ShyftNetwork/go-empyrean/accounts/abi/bind/backends"
	"github.com/ShyftNetwork/go-empyrean/common"
	"github.com/ShyftNetwork/go-empyrean/consensus/ethash"
	"github.com/ShyftNetwork/go-empyrean/core"
	"github.com/ShyftNetwork/go-empyrean/core/types"
	"github.com/ShyftNetwork/go-empyrean/crypto"
	"github.com/ShyftNetwork/go-empyrean/eth"
	"github.com/ShyftNetwork/go-empyrean/shyfttest"
)

// @SHYFT NOTE: test ShyftTracer
const (
	testAddress = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
)

var testKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")

var waitDeployedTests = map[string]struct {
	code        string
	gas         uint64
	wantAddress common.Address
	wantErr     error
}{
	"successful deploy": {
		code:        `6060604052600a8060106000396000f360606040526008565b00`,
		gas:         3000000,
		wantAddress: common.HexToAddress("0x3a220f351252089d385b29beca14e27f204c296a"),
	},
	"empty code": {
		code:        ``,
		gas:         300000,
		wantErr:     bind.ErrNoCodeAfterDeploy,
		wantAddress: common.HexToAddress("0x3a220f351252089d385b29beca14e27f204c296a"),
	},
}

// @SHYFT NOTE: Setup DB for Testing Before Each Test
func TestMain(m *testing.M) {
	shyfttest.PgTestDbSetup()
	retCode := m.Run()
	shyfttest.PgTestTearDown()
	os.Exit(retCode)
}
func TestWaitDeployed(t *testing.T) {
	for name, test := range waitDeployedTests {
		// @SHYFT NOTE: Prepare Test DB & Teardown
		shyfttest.TruncateTables()
		backend := backends.NewSimulatedBackend(core.GenesisAlloc{
			crypto.PubkeyToAddress(testKey.PublicKey): {Balance: big.NewInt(10000000000)},
		})

		// Create the transaction.
		tx := types.NewContractCreation(0, big.NewInt(0), test.gas, big.NewInt(1), common.FromHex(test.code))
		tx, _ = types.SignTx(tx, types.HomesteadSigner{}, testKey)

		// Wait for it to get mined in the background.
		var (
			err     error
			address common.Address
			mined   = make(chan struct{})
			ctx     = context.Background()
		)
		eth.NewShyftTestLDB()
		shyftTracer := new(eth.ShyftTracer)

		ethConf := &eth.Config{
			Genesis:   core.DeveloperGenesisBlock(15, common.Address{}),
			Etherbase: common.HexToAddress(testAddress),
			Ethash: ethash.Config{
				PowMode: ethash.ModeTest,
			},
		}

		eth.SetGlobalConfig(ethConf)
		eth.InitTracerEnv()
		go func() {
			core.SetIShyftTracer(shyftTracer)

			address, err = bind.WaitDeployed(ctx, backend, tx)
			close(mined)
		}()

		// Send and mine the transaction.
		backend.SendTransaction(ctx, tx)
		backend.Commit()

		select {
		case <-mined:
			if err != test.wantErr {
				t.Errorf("test %q: error mismatch: got %q, want %q", name, err, test.wantErr)
			}
			if address != test.wantAddress {
				t.Errorf("test %q: unexpected contract address %s", name, address.Hex())
			}
		case <-time.After(2 * time.Second):
			t.Errorf("test %q: timeout", name)
		}
	}
}
