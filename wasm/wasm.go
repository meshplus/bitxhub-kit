package wasm

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/meshplus/bitxhub-kit/types"
	"github.com/meshplus/bitxhub-model/pb"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

const (
	CONTEXT_ARGMAP    = "argmap"
	CONTEXT_INTERFACE = "interface"
)

var (
	errorLackOfMethod = fmt.Errorf("wasm execute: lack of method name")
)

func getInstance(code []byte, imports *wasmer.Imports, instances map[string]wasmer.Instance) (wasmer.Instance, error) {
	ret := sha256.Sum256(code)
	v, ok := instances[string(ret[:])]
	if ok {
		return v, nil
	}

	instance, err := wasmer.NewInstanceWithImports(code, imports)
	if err != nil {
		return wasmer.Instance{}, err
	}

	instances[string(ret[:])] = instance

	return instance, nil
}

// Wasm represents the wasm vm in BitXHub
type Wasm struct {
	// wasm instance
	Instance wasmer.Instance

	context map[string]interface{}
	argMap  map[int]int
}

// Contract represents the smart contract structure used in the wasm vm
type Contract struct {
	// contract byte
	Code []byte

	// contract hash
	Hash types.Hash
}

// New creates a wasm vm instance
func New(contractByte []byte, imports *wasmer.Imports, instances map[string]wasmer.Instance) (*Wasm, error) {
	wasm := &Wasm{}

	contract := &Contract{}
	if err := json.Unmarshal(contractByte, contract); err != nil {
		return wasm, fmt.Errorf("contract byte not correct")
	}

	if len(contract.Code) == 0 {
		return wasm, fmt.Errorf("contract byte is empty")
	}

	instance, err := getInstance(contract.Code, imports, instances)
	if err != nil {
		return nil, err
	}

	wasm.Instance = instance
	wasm.argMap = make(map[int]int)

	return wasm, nil
}

func EmptyImports() (*wasmer.Imports, error) {
	return wasmer.NewImports(), nil
}

func (w *Wasm) Execute(input []byte) ([]byte, error) {
	payload := &pb.InvokePayload{}
	if err := proto.Unmarshal(input, payload); err != nil {
		return nil, err
	}

	if payload.Method == "" {
		return nil, errorLackOfMethod
	}

	methodName, ok := w.Instance.Exports[payload.Method]
	if !ok {
		return nil, fmt.Errorf("wrong rule contract")
	}
	slice := make([]interface{}, len(payload.Args))
	for i := range slice {
		arg := payload.Args[i]
		switch arg.Type {
		case pb.Arg_I32:
			temp, err := strconv.Atoi(string(arg.Value))
			if err != nil {
				return nil, err
			}
			slice[i] = temp
		case pb.Arg_I64:
			temp, err := strconv.ParseInt(string(arg.Value), 10, 64)
			if err != nil {
				return nil, err
			}
			slice[i] = temp
		case pb.Arg_F32:
			temp, err := strconv.ParseFloat(string(arg.Value), 32)
			if err != nil {
				return nil, err
			}
			slice[i] = temp
		case pb.Arg_F64:
			temp, err := strconv.ParseFloat(string(arg.Value), 64)
			if err != nil {
				return nil, err
			}
			slice[i] = temp
		case pb.Arg_String:
			inputPointer, err := w.SetString(string(arg.Value))
			if err != nil {
				return nil, err
			}
			slice[i] = inputPointer
		case pb.Arg_Bytes:
			inputPointer, err := w.SetBytes(arg.Value)
			if err != nil {
				return nil, err
			}
			slice[i] = inputPointer
		case pb.Arg_Bool:
			inputPointer, err := strconv.Atoi(string(arg.Value))
			if err != nil {
				return nil, err
			}
			slice[i] = inputPointer
		default:
			return nil, fmt.Errorf("input type not support")
		}
	}

	w.context[CONTEXT_ARGMAP] = w.argMap
	w.Instance.SetContextData(w.context)

	result, err := methodName(slice...)
	if err != nil {
		return nil, err
	}

	return []byte(result.String()), err
}
