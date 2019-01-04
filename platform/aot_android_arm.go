package platform

/*
#include "vm_def.h"
*/
import "C"

import (
	"github.com/perlin-network/life/exec"
	"unsafe"
)

type AOTContext struct {
	dlHandle unsafe.Pointer
	vmHandle *C.struct_VirtualMachine
}

func (c *AOTContext) UnsafeInvokeFunction_0(vm *exec.VirtualMachine, name string) uint64 {
	return 0
}

func (c *AOTContext) UnsafeInvokeFunction_1(vm *exec.VirtualMachine, name string, p0 uint64) uint64 {
	return 0
}

func (c *AOTContext) UnsafeInvokeFunction_2(vm *exec.VirtualMachine, name string, p0, p1 uint64) uint64 {
	return 0
}

func FullAOTCompile(vm *exec.VirtualMachine) *AOTContext {
	return nil
}