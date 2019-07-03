package gofrida

//#include "frida-core.h"
import "C"


const(
	FRIDA_SCRIPT_RUNTIME_DEFAULT=C.FRIDA_SCRIPT_RUNTIME_DEFAULT
	FRIDA_SCRIPT_RUNTIME_DUK=C.FRIDA_SCRIPT_RUNTIME_DEFAULT
	FRIDA_SCRIPT_RUNTIME_V8=C.FRIDA_SCRIPT_RUNTIME_DEFAULT
)
type FridaScriptOptions struct{
	ptr uintptr
}
func FridaScriptOptionsFormPtr(_ptr uintptr)*FridaScriptOptions{
	return &FridaScriptOptions{_ptr}
}
func (this* FridaScriptOptions)CPtr()uintptr{
	return this.ptr
}
