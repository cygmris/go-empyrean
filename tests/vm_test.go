// Copyright 2014 The go-ethereum Authors
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

package tests

import (
	"os"
	"testing"

	"github.com/ShyftNetwork/go-empyrean/common"
	"github.com/ShyftNetwork/go-empyrean/consensus/ethash"
	"github.com/ShyftNetwork/go-empyrean/core"
	"github.com/ShyftNetwork/go-empyrean/core/vm"
	"github.com/ShyftNetwork/go-empyrean/eth"
	"github.com/ShyftNetwork/go-empyrean/shyfttest"
	"github.com/docker/docker/pkg/reexec"
)

func TestMain(m *testing.M) {
	// Reset Pg DB
	shyfttest.PgTestDbSetup()
	// check if we have been reexec'd

	if reexec.Init() {
		return
	}
	retCode := m.Run()
	shyfttest.PgTestTearDown()
	os.Exit(retCode)
}

func TestVM(t *testing.T) {
	t.Parallel()
	//@SHYFT //SETS UP OUR TEST ENV - Parallel tests creating issues so the below may not work with parallel
	eth.NewShyftTestLDB()
	shyftTracer := new(eth.ShyftTracer)
	core.SetIShyftTracer(shyftTracer)

	ethConf := &eth.Config{
		Genesis:   core.DeveloperGenesisBlock(15, common.Address{}),
		Etherbase: common.HexToAddress(testAddress),
		Ethash: ethash.Config{
			PowMode: ethash.ModeTest,
		},
	}

	eth.SetGlobalConfig(ethConf)
	eth.InitTracerEnv()
	vmt := new(testMatcher)
	vmt.fails("^vmSystemOperationsTest.json/createNameRegistrator$", "fails without parallel execution")

	vmt.skipLoad(`^vmInputLimits(Light)?.json`) // log format broken

	vmt.skipShortMode("^vmPerformanceTest.json")
	vmt.skipShortMode("^vmInputLimits(Light)?.json")

	vmt.walk(t, vmTestDir, func(t *testing.T, name string, test *VMTest) {
		withTrace(t, test.json.Exec.GasLimit, func(vmconfig vm.Config) error {
			return vmt.checkFailure(t, name, test.Run(vmconfig))
		})
	})
}
