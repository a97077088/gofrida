package gofrida

/*
#include <stdlib.h>
 */
import "C"
import "unsafe"


type FridaSpawnOptions struct{
	ptr uintptr
}
func FridaSpawnOptionsFormPtr(_ptr uintptr)*FridaSpawnOptions{
	return &FridaSpawnOptions{_ptr}
}
func (this* FridaSpawnOptions)CPtr()uintptr{
	return this.ptr
}
func (this* FridaSpawnOptions)Set_argv(args []string) {
	arg := make([](*C.char), 0)  //C语言char*指针创建切片
	l := len(args)
	for i,_ := range args{
		char := C.CString(args[i])
		defer C.free(unsafe.Pointer(char)) //释放内存
		strptr := (*C.char)(unsafe.Pointer(char))
		arg = append(arg, strptr)  //将char*指针加入到arg切片
	}
	n:= uintptr((unsafe.Pointer(&arg[0])))
	frida_spawn_options_set_argv(this,n,l)
}
