package wasm

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/xunleichain/tc-wasm/mock/log"
	"github.com/xunleichain/tc-wasm/mock/state"
	"github.com/xunleichain/tc-wasm/mock/types"
	"github.com/xunleichain/tc-wasm/vm"
)

var (
	cState  *state.StateDB
	cAddr   types.Address
	ctxTime uint64
)

func init() {
	ctxTime = 1565078742
	cAddr = types.BytesToAddress([]byte{1})

	cState, _ = state.New()
	cState.AddBalance(cAddr, big.NewInt(int64(10000)))
}

func TestParseInitArgs(t *testing.T) {
	var data []byte
	data = append(data, vm.WasmBytes...)
	data = append(data, []byte("XLTC")...)

	args := "{\"num\": 100, \"name\":\"xxxx\"}"
	argsLen := uint16(len(args))
	argsBuf := bytes.NewBuffer([]byte{})
	if err := binary.Write(argsBuf, binary.BigEndian, argsLen); err != nil {
		t.Fatalf("binary.Write fail: %s", err)
	}

	data = append(data, argsBuf.Bytes()...)
	data = append(data, []byte(args)...)

	code := "HelloWorld"
	data = append(data, []byte(code)...)

	tmpInput, tmpCode, err := vm.ParseInitArgsAndCode(data)
	if err != nil {
		t.Fatalf("ParseInitArgsAndCode fail: %s", err)
	}

	t.Logf("input: %s", string(tmpInput))
	t.Logf("code: %s", string(tmpCode))
	if !bytes.HasPrefix(tmpInput, []byte("Init|")) {
		t.Fatalf("input with prefix(Init|): %s", string(tmpInput))
	}
	if !bytes.Equal(tmpInput[5:], []byte(args)) {
		t.Fatalf("input not match: wanted(%s), got(%s)", args, string(tmpInput[5:]))
	}
	if !bytes.Equal([]byte(code), tmpCode) {
		t.Fatalf("code not match: wanted(%s), got(%s)", string(code), string(tmpCode))
	}
}

func TestCallContract(t *testing.T) {
	wasmContractFile1 := "../../../testdata/contract.wasm"
	contractCode1, err := ioutil.ReadFile(wasmContractFile1)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}

	wasmContractFile2 := "../../../testdata/contract1.wasm"
	contractCode2, err := ioutil.ReadFile(wasmContractFile2)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}

	wasmContractFile3 := "../../../testdata/contract2.wasm"
	contractCode3, err := ioutil.ReadFile(wasmContractFile3)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}

	addr1 := types.BytesToAddress([]byte{114})
	cState.SetCode(addr1, contractCode1)

	addr2 := types.BytesToAddress([]byte{115})
	cState.SetCode(addr2, contractCode2)

	addr3 := types.BytesToAddress([]byte{116})
	cState.SetCode(addr3, contractCode3)

	contract := vm.NewContract(cAddr.Bytes(), addr1.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr1
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr1,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 1000000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr1.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	action := "none"
	params := "{\"contract1\":\"0x0000000000000000000000000000000000000073\",\"contract2\":\"0x0000000000000000000000000000000000000074\"}"
	input := make([]byte, len(action)+len(params)+5)
	copy(input[0:], vm.WasmBytes[0:4])
	copy(input[4:], action)
	copy(input[4+len(action):], []byte{'|'})
	copy(input[5+len(action):], params)
	eng.Contract.Input = input[4:]
	ret, err := eng.Run(app, input)
	t.Logf("ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestNotify(t *testing.T) {
	wasmFile := "../../../testdata/notify.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{113})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	logs := cState.Logs()
	for i := 0; i < len(logs); i++ {
		t.Logf("log %d %s", i, logs[i].String())
	}
	return
}

func TestToken(t *testing.T) {
	wasmFile := "../../../testdata/token.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{112})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("strlen ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestMalloc(t *testing.T) {
	wasmFile := "../../../testdata/malloc.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{111})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("malloc ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestPrints(t *testing.T) {
	wasmFile := "../../../testdata/prints.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{110})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("prints ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestLog(t *testing.T) {
	wasmFile := "../../../testdata/log.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{109})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("log ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	logs := cState.Logs()
	for i := 0; i < len(logs); i++ {
		t.Logf("log %d %s", i, logs[i].String())
	}
	return
}

func TestSelfDestruct(t *testing.T) {
	wasmFile := "../../../testdata/selfdestruct.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{108})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	t.Logf("from account balance: %d before exec contract method", cState.GetBalance(addr))
	t.Logf("to account balance: %d before exec contract method", cState.GetBalance(types.HexToAddress("0x0000000000000000000000000000000000000001")))
	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	t.Logf("from account code: 0x%x before exec contract method", cState.GetCode(addr))
	t.Logf("from account cache code: %v before exec contract method", eng.AppByName(addr.String()))

	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)

	t.Logf("selfdestruct ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	t.Logf("from account balance: %d after exec contract method", cState.GetBalance(addr))
	t.Logf("to account balance: %d after exec contract method", cState.GetBalance(types.HexToAddress("0x0000000000000000000000000000000000000001")))
	t.Logf("from account code: 0x%x after exec contract method", cState.GetCode(addr))
	t.Logf("from account cache code: %v after exec contract method", eng.AppByName(addr.String()))
	return
}

func TestSelfAddress(t *testing.T) {
	wasmFile := "../../../testdata/selfaddress.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{107})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("getSelfAddress ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestGetBalance(t *testing.T) {
	wasmFile := "../../../testdata/getbalance.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{106})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("getBalance ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestEcrecover(t *testing.T) {
	wasmFile := "../../../testdata/ecrecover.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{105})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("ecrecover ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestRipemd160(t *testing.T) {
	wasmFile := "../../../testdata/ripemd160.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{103})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("ripemd160 ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestSha256(t *testing.T) {
	wasmFile := "../../../testdata/sha256.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{102})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("sha256 ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestKeccak256(t *testing.T) {
	wasmFile := "../../../testdata/keccak256.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{101})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("keccak256 ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	return
}

func TestTransfer(t *testing.T) {
	wasmFile := "../../../testdata/transfer.wasm"
	code, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Logf("read wasm code fail: %v", err)
		return
	}
	addr := types.BytesToAddress([]byte{100})
	cState.AddBalance(addr, big.NewInt(int64(10000)))
	cState.SetCode(addr, code)

	t.Logf("from account balance: %d before exec contract method", cState.GetBalance(addr))
	t.Logf("to account balance: %d before exec contract method", cState.GetBalance(types.HexToAddress("0x0000000000000000000000000000000000000001")))
	contract := vm.NewContract(cAddr.Bytes(), addr.Bytes(), big.NewInt(100), 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 100000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Logf("new app fail: err: %v", err)
		return
	}
	input := []byte{0x00, 0x61, 0x73, 0x6d, 'a', '|', 'a'}
	ret, err := eng.Run(app, input)
	t.Logf("transfer ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())
	t.Logf("from account balance: %d after exec contract method", cState.GetBalance(addr))
	t.Logf("to account balance: %d after exec contract method", cState.GetBalance(types.HexToAddress("0x0000000000000000000000000000000000000001")))
	return
}

func TestDataCall(t *testing.T) {
	wasmFile := "/Users/stephen/dev/workspaces/blockchain/xunlei/tc-wasm/src/github.com/xunleichain/tc-wasm/testdata/data.wasm"
	wasmBytes, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		t.Fatalf("read module file %s failed: %v", wasmFile, err)
	}

	addr := types.BytesToAddress([]byte{0x1, 0x2, 0x3, 0x4})
	cState.SetCode(addr, wasmBytes)

	caller := types.BytesToAddress([]byte{0x4, 0x3, 0x2, 0x1})
	contract := vm.NewContract(caller.Bytes(), addr.Bytes(), nil, 0)
	contract.CodeAddr = &addr
	ctx := Context{
		Time:        new(big.Int).SetUint64(ctxTime),
		Token:       addr,
		BlockNumber: big.NewInt(3456),
	}
	eng := vm.NewEngine(contract, 1000000, cState, log.Test())
	Inject(&ctx, cState)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		t.Fatalf("new app error: %v", err)
	}

	input := make([]byte, 0)
	eng.Contract.Input = input
	ret, err := eng.Run(app, input)
	t.Logf("ret: %d, err: %v", ret, err)
	t.Logf("gas used: %d", eng.GasUsed())

	// m, err := wasm.ReadModule(bytes.NewReader(wasmBytes), vm.NewEnvTable().Resolver)
	// if err != nil {
	// 	t.Fatalf("new module failed: %v", err)
	// }
	// vm, err := exec.NewVM(m, nil)
	// if err != nil {
	// 	t.Fatalf("new vm failed: %v", err)
	// }
	// _, err = vm.ExecCode(2)
	// if err != nil {
	// 	t.Fatalf("run error: %v", err)
	// }
}
