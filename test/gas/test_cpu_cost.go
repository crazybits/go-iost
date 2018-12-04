package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/db"

	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
	v8 "github.com/iost-official/go-iost/vm/v8vm"
)

var (
	OpList = map[string][]string{
		"base": []string{
			"Empty",
			"Call",
			"New",
			"Throw",
			"Yield",
			"Member",
			"Meta",
			"Assignment",
			"Plus",
			"Add",
			"Sub",
			"Mutiple",
			"Div",
			"Not",
			"And",
			"Conditional",
		},
		"lib": []string{
			"StringToString",
			"StringValueOf",
			"StringConcat",
			"StringIncludes",
			"StringEndsWith",
			"StringIndexOf",
			"StringLastIndexOf",
			"StringReplace",
			"StringSearch",
			"StringSplit",
			"StringStartsWith",
			"StringSlice",
			"StringToLowerCase",
			"StringToUpperCase",
			"StringTrim",
			"StringTrimLeft",
			"StringTrimRight",
			"StringRepeat",
			"ArrayIsArray",
			"ArrayOf",
			"ArrayConcat",
			"ArrayEvery",
			"ArrayFilter",
			"ArrayFind",
			"ArrayFindIndex",
			"ArrayForEach",
			"ArrayIncludes",
			"ArrayIndexOf",
			"ArrayJoin",
			"ArrayKeys",
			"ArrayLastIndexOf",
			"ArrayMap",
			"ArrayPop",
			"ArrayPush",
			"ArrayReverse",
			"ArrayShift",
			"ArraySlice",
			"ArraySort",
			"ArraySplice",
			"ArrayToString",
			"ArrayUnshift",
			"JSONParse",
			"JSONStringify",
		},
	}
)

var vmPool *v8.VMPool
var testDataPath = "./test_data/"
var BaseCPUCost = int64(2000)

func RunOp(vi *database.Visitor, name string, api string, num int) (float64, int64) {
	b, err := ioutil.ReadFile(path.Join(testDataPath, name))
	if err != nil {
		log.Fatalf("Read file failed: %v", err)
	}
	code := string(b)

	now := time.Now()

	ctx := host.NewContext(nil)
	ctx.Set("gas_price", int64(1))
	ctx.GSet("gas_limit", int64(100000000))
	ctx.Set("contract_name", name)

	host := host.NewHost(ctx, vi, nil, ilog.DefaultLogger())
	expTime := time.Now().Add(time.Second * 10)
	host.SetDeadline(expTime)

	contract := &contract.Contract{
		ID:   name,
		Code: code,
	}

	contract.Code, err = vmPool.Compile(contract)
	if err != nil {
		log.Fatalf("Compile contract failed: %v", err)
	}

	_, cost, err := vmPool.LoadAndCall(host, contract, api, num)

	if err != nil {
		log.Fatalf("LoadAndCall %v.%v %v failed: %v", contract, api, num, err)
	}

	return time.Now().Sub(now).Seconds(), BaseCPUCost + cost.CPU
}

func init() {
	// TODO The number of pool need adjust
	vmPool = v8.NewVMPool(30, 30)
	vmPool.Init()
}

func main() {
	mvccdb, err := db.NewMVCCDB("mvccdb")
	if err != nil {
		log.Fatalf("New MVCC DB failed: %v", err)
	}
	vi := database.NewVisitor(100, mvccdb)

	for _, opType := range []string{"base", "lib"} {
		for _, op := range OpList[opType] {
			fmt.Printf("========================%v========================\n", op)
			tcost, ccost := RunOp(
				vi,
				fmt.Sprintf("%v_op.js", opType),
				fmt.Sprintf("do%v", op),
				0,
			)
			fmt.Printf("Time: %0.3fs\n", tcost)
			fmt.Printf("CPU Cost: %vgas\n", ccost)
		}
	}

	os.RemoveAll("mvccdb")
}