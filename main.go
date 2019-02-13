package main

import (
	"flag"
	"fmt"
	"github.com/fendouhyz/life/exec"
	"github.com/fendouhyz/life/platform"
	"io/ioutil"
	"time"
)

// Resolver defines imports for WebAssembly modules ran in Life.
type Resolver struct {
	tempRet0 int64
}

// ResolveFunc defines a set of import functions that may be called within a WebAssembly module.
/*
	定义了一个入口函数的集合，这些函数被调用的时候，会加载进一个WebAssembly module里面
	1. 未来在此加入访问block数据的api
	2. 未来还需要扩展一些string等高级数据类型，类似本体的做法，（不确定是否在这里扩展）
*/
func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	fmt.Printf("Resolve func: %s %s\n", module, field)
	switch module {
	case "env":
		switch field {
		case "__life_ping":
			return func(vm *exec.VirtualMachine) int64 {
				return vm.GetCurrentFrame().Locals[0] + 1
			}
		case "__life_log":
			return func(vm *exec.VirtualMachine) int64 {
				ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				msgLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				msg := vm.Memory[ptr : ptr+msgLen]
				fmt.Printf("[app] %s\n", string(msg))
				return 0
			}
		case "print_i64":
			return func(vm *exec.VirtualMachine) int64 {
				fmt.Printf("[app] print_i64: %d\n", vm.GetCurrentFrame().Locals[0])
				return 0
			}

		default:
			panic(fmt.Errorf("unknown field: %s", field))
		}
	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

// ResolveGlobal defines a set of global variables for use within a WebAssembly module.
func (r *Resolver) ResolveGlobal(module, field string) int64 {
	fmt.Printf("Resolve global: %s %s\n", module, field)
	switch module {
	case "env":
		switch field {
		case "__life_magic":
			return 424
		default:
			panic(fmt.Errorf("unknown field: %s", field))
		}
	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

func main() {
	entryFunctionFlag := flag.String("entry", "app_main", "entry function name")
	pmFlag := flag.Bool("polymerase", false, "enable the Polymerase engine")
	noFloatingPointFlag := flag.Bool("no-fp", false, "disable floating point")
	flag.Parse()



	// Read WebAssembly *.wasm file.
	//input, err := ioutil.ReadFile(flag.Arg(0))
	/*
		这里注释和原读入文件，为了测试方便
	*/
	input, err := ioutil.ReadFile("./tests/imports.wasm")
	if err != nil {
		panic(err)
	}

	// Instantiate a new WebAssembly VM with a few resolved imports.
	//实例化一个虚拟机
	vm, err := exec.NewVirtualMachine(input, exec.VMConfig{
		DefaultMemoryPages:   128,
		DefaultTableSize:     65536,
		DisableFloatingPoint: *noFloatingPointFlag,  //这里目前还不太动这个flag的意思，字面意思是漂流点？浮点指针？指向浮点数的指针？
	}, new(Resolver), nil)

	if err != nil {
		panic(err)
	}

	/*
		pmFlag是polymerase的开关，命令行读取，具体什么意思？字面意思是聚合酶
		AOTCompile以及AOTService又是什么意思呢？
	*/
	if *pmFlag {
		compileStartTime := time.Now()
		fmt.Println("[Polymerase] Compilation started.")
		aotSvc := platform.FullAOTCompile(vm)
		if aotSvc != nil {
			compileEndTime := time.Now()
			fmt.Printf("[Polymerase] Compilation finished successfully in %+v.\n", compileEndTime.Sub(compileStartTime))
			vm.SetAOTService(aotSvc)
		} else {
			fmt.Println("[Polymerase] The current platform is not yet supported.")
		}
	}

	// Get the function ID of the entry function to be executed.
	/*
		1.entryFunctionFlag是从命令行获取的，如果没有读到，自动命令entryID为0
		2.这里是为了方便自定义入口函数
	*/
	entryID, ok := vm.GetFunctionExport(*entryFunctionFlag)
	if !ok {
		fmt.Printf("Entry function %s not found; starting from 0.\n", *entryFunctionFlag)
		entryID = 0
	}

	start := time.Now()

	// If any function prior to the entry function was declared to be
	// called by the module, run it first.
	/*
		在这个module里面，如果有优先级高于入口函数的function，先运行他
		！注意PrintStackTrace这个函数，调试很有用
	*/
	if vm.Module.Base.Start != nil {
		startID := int(vm.Module.Base.Start.Index)
		_, err := vm.Run(startID)
		if err != nil {
			vm.PrintStackTrace()
			panic(err)
		}
	}
	var args []int64
	/*这里注释了原有的代码，因为不直接从命令行读入了*/
	//for _, arg := range flag.Args()[1:] {
	//	fmt.Println(arg)
	//	if ia, err := strconv.Atoi(arg); err != nil {
	//		panic(err)
	//	} else {
	//		args = append(args, int64(ia))
	//	}
	//}

	// Run the WebAssembly module's entry function.
	ret, err := vm.Run(entryID, args...)
	if err != nil {
		vm.PrintStackTrace()
		panic(err)
	}
	end := time.Now()

	fmt.Printf("return value = %d, duration = %v\n", ret, end.Sub(start))
}
