package gofrida

import (
	"errors"
	"io/ioutil"
)

type FridaSession struct {
	ptr uintptr
}

func FridaSessionFormPtr(_ptr uintptr) *FridaSession {
	return &FridaSession{_ptr}
}
func (this *FridaSession) CPtr() uintptr {
	return this.ptr
}

func (this *FridaSession) Detach() {
	frida_session_detach_sync(this, nil, nil)
	frida_unref(this.CPtr())
}

func (this *FridaSession) Create_Script_with_path(_scriptpath string) (*FridaScript, error) {
	bt, err := ioutil.ReadFile(_scriptpath)
	if err != nil {
		return nil, err
	}
	return this.Create_Script_with_name_script(_scriptpath, string(bt))
}
func (this *FridaSession) Create_Script_with_name_script(_scriptname string, _script string) (*FridaScript, error) {
	var gerr *GError
	opt := frida_script_options_new()
	frida_script_options_set_name(opt, _scriptname)
	frida_script_options_set_runtime(opt, FRIDA_SCRIPT_RUNTIME_V8)
	defer g_clear_object(&opt.ptr) //暂时不知道咋释放

	sc := frida_session_create_script_sync(this, _script, opt, nil, &gerr)
	if gerr != nil {
		return nil, errors.New(gerr.Message())
	}
	return sc, nil
}
