package vm

import (
	"fmt"

	"github.com/go-interpreter/wagon/wasm"
)

type EnvFunc interface {
	Call(index int64, ops interface{}, args []uint64) (uint64, error)
	Gas(index int64, ops interface{}, args []uint64) (uint64, error)
}

// EnvTable stand for env's info which we will register for wasm module before it run.
type EnvTable struct {
	Exports         wasm.SectionExports
	Module          wasm.Module
	importFuncCnt   uint32
	importGlobalCnt uint32
}

var (
	gEnvTable *EnvTable
)

func init() {
	env := EnvTable{
		Exports: wasm.SectionExports{
			Entries: make(map[string]wasm.ExportEntry),
			Names:   make([]string, 0),
		},
	}
	env.Module = wasm.Module{
		Export:             &env.Exports,
		FunctionIndexSpace: make([]wasm.Function, 0),
		GlobalIndexSpace:   make([]wasm.GlobalEntry, 0),
	}

	gEnvTable = &env

	gEnvTable.RegisterFunc("TC_CallContract", new(TCCallContract))
	gEnvTable.RegisterFunc("TC_DelegateCallContract", new(TCDelegateCallContract))

	gEnvTable.RegisterFunc("TC_BigIntAdd", new(TCBigIntAdd))
	gEnvTable.RegisterFunc("TC_BigIntSub", new(TCBigIntSub))
	gEnvTable.RegisterFunc("TC_BigIntMul", new(TCBigIntMul))
	gEnvTable.RegisterFunc("TC_BigIntDiv", new(TCBigIntDiv))
	gEnvTable.RegisterFunc("TC_BigIntMod", new(TCBigIntMod))
	gEnvTable.RegisterFunc("TC_BigIntCmp", new(TCBigIntCmp))
	gEnvTable.RegisterFunc("TC_BigIntToInt64", new(TCBigIntToInt64))

	gEnvTable.RegisterFunc("exit", new(TCExit))
	gEnvTable.RegisterFunc("abort", new(TCAbort))
	gEnvTable.RegisterFunc("malloc", new(TCMalloc))
	gEnvTable.RegisterFunc("calloc", new(TCCalloc))
	gEnvTable.RegisterFunc("realloc", new(TCRealloc))
	gEnvTable.RegisterFunc("prints_l", new(TCPrintsl))
	gEnvTable.RegisterFunc("free", new(TCFree))
	gEnvTable.RegisterFunc("memcpy", new(TCMemcpy))
	gEnvTable.RegisterFunc("memset", new(TCMemset))
	gEnvTable.RegisterFunc("memmove", new(TCMemmove))
	gEnvTable.RegisterFunc("memcmp", new(TCMemcmp))
	gEnvTable.RegisterFunc("strcmp", new(TCStrcmp))
	gEnvTable.RegisterFunc("strcpy", new(TCStrcpy))
	gEnvTable.RegisterFunc("strlen", new(TCStrlen))
	gEnvTable.RegisterFunc("strconcat", new(TCStrconcat))
	gEnvTable.RegisterFunc("atoi", new(TCAtoi))
	gEnvTable.RegisterFunc("atoi64", new(TCAtoi64))
	//	gEnvTable.RegisterFunc("atof32", new(TCAtof32))
	//	gEnvTable.RegisterFunc("atof64", new(TCAtof64))
	gEnvTable.RegisterFunc("itoa", new(TCItoa))
	gEnvTable.RegisterFunc("i64toa", new(TCI64toa))

	gEnvTable.RegisterFunc("TC_GetMsgData", new(TCGetMsgData))
	gEnvTable.RegisterFunc("TC_GetMsgGas", new(TCGetMsgGas))
	gEnvTable.RegisterFunc("TC_GetMsgSender", new(TCGetMsgSender))
	gEnvTable.RegisterFunc("TC_GetMsgSign", new(TCGetMsgSign))
	gEnvTable.RegisterFunc("TC_Assert", new(TCAssert))
	gEnvTable.RegisterFunc("TC_Require", new(TCRequire))
	gEnvTable.RegisterFunc("TC_GasLeft", new(TCGasLeft))
	gEnvTable.RegisterFunc("TC_RequireWithMsg", new(TCRequireWithMsg))
	gEnvTable.RegisterFunc("TC_Revert", new(TCRevert))
	gEnvTable.RegisterFunc("TC_RevertWithMsg", new(TCRevertWithMsg))
	gEnvTable.RegisterFunc("TC_IsHexAddress", new(TCIsHexAddress))
	gEnvTable.RegisterFunc("TC_Payable", new(TCPayable))

	gEnvTable.RegisterFunc("TC_Prints", new(TCPrints))
	gEnvTable.RegisterFunc("TC_GetSelfAddress", new(TCGetSelfAddress))
	gEnvTable.RegisterFunc("TC_Ripemd160", new(TCRipemd160))
	gEnvTable.RegisterFunc("TC_Sha256", new(TCSha256))
	gEnvTable.RegisterFunc("TC_Keccak256", new(TCKeccak256))

	// go json api (optional)
	gEnvTable.RegisterFunc("TC_JsonParse", new(TCJSONParse))
	gEnvTable.RegisterFunc("TC_JsonGetInt", new(TCJSONGetInt))
	gEnvTable.RegisterFunc("TC_JsonGetInt64", new(TCJSONGetInt64))
	gEnvTable.RegisterFunc("TC_JsonGetString", new(TCJSONGetString))
	gEnvTable.RegisterFunc("TC_JsonGetAddress", new(TCJSONGetAddress))
	gEnvTable.RegisterFunc("TC_JsonGetBigInt", new(TCJSONGetBigInt))
	gEnvTable.RegisterFunc("TC_JsonGetFloat", new(TCJSONGetFloat))
	gEnvTable.RegisterFunc("TC_JsonGetDouble", new(TCJSONGetDouble))
	gEnvTable.RegisterFunc("TC_JsonGetObject", new(TCJSONGetObject))
	gEnvTable.RegisterFunc("TC_JsonNewObject", new(TCJSONNewObject))
	gEnvTable.RegisterFunc("TC_JsonPutInt", new(TCJSONPutInt))
	gEnvTable.RegisterFunc("TC_JsonPutInt64", new(TCJSONPutInt64))
	gEnvTable.RegisterFunc("TC_JsonPutString", new(TCJSONPutString))
	gEnvTable.RegisterFunc("TC_JsonPutAddress", new(TCJSONPutAddress))
	gEnvTable.RegisterFunc("TC_JsonPutBigInt", new(TCJSONPutBigInt))
	gEnvTable.RegisterFunc("TC_JsonPutFloat", new(TCJSONPutFloat))
	gEnvTable.RegisterFunc("TC_JsonPutDouble", new(TCJSONPutDouble))
	gEnvTable.RegisterFunc("TC_JsonPutObject", new(TCJSONPutObject))
	gEnvTable.RegisterFunc("TC_JsonToString", new(TCJSONToString))
}

// NewEnvTable new EnvTable
func NewEnvTable() *EnvTable {
	return gEnvTable
}

func (env *EnvTable) resolveImport(name string) (*wasm.Module, error) {
	if name != "env" {
		panic(fmt.Sprintf("invalid name: %s", name))
	}
	return &env.Module, nil
}

func (env *EnvTable) Resolver(name string) (*wasm.Module, error) {
	return env.resolveImport(name)
}

// RegisterFunc Register env function for wasm module
func (env *EnvTable) RegisterFunc(name string, fn EnvFunc) {
	if entry, exist := env.Exports.Entries[name]; exist {
		env.Module.FunctionIndexSpace[entry.Index].Host = fn
		return
	}

	env.Exports.Names = append(env.Exports.Names, name)
	env.Exports.Entries[name] = wasm.ExportEntry{
		FieldStr: name,
		Kind:     wasm.ExternalFunction,
		Index:    env.importFuncCnt,
	}
	env.Module.FunctionIndexSpace = append(env.Module.FunctionIndexSpace, wasm.Function{
		Sig:  &wasm.FunctionSig{},
		Body: &wasm.FunctionBody{Module: &env.Module},
		Host: fn,
		Name: name,
	})
	env.importFuncCnt++
}

// RegisterGlobal Register env global for wasm module
func (env *EnvTable) RegisterGlobal(name string, v interface{}) {
	env.Exports.Names = append(env.Exports.Names, name)
	env.Exports.Entries[name] = wasm.ExportEntry{
		FieldStr: name,
		Kind:     wasm.ExternalGlobal,
		Index:    env.importGlobalCnt,
	}
	env.Module.GlobalIndexSpace = append(env.Module.GlobalIndexSpace, wasm.GlobalEntry{
		Type: wasm.GlobalVar{
			Type: wasm.ValueTypeI32,
		},
		Init: []byte{65, 9, 11}, // for test.
	})
	env.importGlobalCnt++
}

// GetFuncByName Get env function by name
func (env *EnvTable) GetFuncByName(name string) EnvFunc {
	if entry, exist := env.Exports.Entries[name]; exist {
		return env.Module.FunctionIndexSpace[entry.Index].Host.(EnvFunc)
	}
	return nil
}
