package gofrida


import (
	"errors"
	"strings"
	"time"
)

const (
	FRIDA_DEVICE_TYPE_LOCAL = iota
	FRIDA_DEVICE_TYPE_REMOTE
	FRIDA_DEVICE_TYPE_USB
)

type FridaDevice struct {
	ptr  uintptr
	Name string
	Id   string
	Type FridaDeviceType
}

func FridaDeviceFormPtr(_ptr uintptr) *FridaDevice {
	return &FridaDevice{ptr: _ptr}
}
func (this *FridaDevice) CPtr() uintptr {
	return this.ptr
}

func (this *FridaDevice) Get_Application_with_Identifier(_Identifier string) (*FridaApplication, error) {
	as, err := this.EnumApplication()
	if err != nil {
		return nil, err
	}
	for _, it := range as {
		if it.Identifier == _Identifier {
			return it, nil
		}
	}
	return nil, errors.New("找不到这个application")
}
func (this *FridaDevice) Get_Process_with_name(_name string,timeout time.Duration) (*FridaProcess, error) {
	var gerr *GError
	p := frida_device_get_process_by_name_sync(this, _name, int(timeout.Milliseconds()), nil, &gerr)
	if gerr != nil {
		return nil, errors.New(gerr.Message())
	}
	p.Fill()
	return p, nil
}

func (this *FridaDevice) Kill_with_name(_name string) error {
	p, err := this.Get_Process_with_name(_name,time.Second*1)
	if err != nil {
		return err
	}
	return this.Kill(p.Pid)
}

func (this *FridaDevice) Kill(_pid int) error {
	var gerr *GError
	frida_device_kill_sync(this, _pid, nil, &gerr)
	if gerr != nil {
		return errors.New(gerr.Message())
	}
	return nil
}

func (this *FridaDevice) EnumProcess() ([]*FridaProcess, error) {
	var gerr *GError
	processlist := frida_device_enumerate_processes_sync(this, nil, &gerr)
	if gerr != nil {
		return nil, errors.New(gerr.Message())
	}
	defer func() {
		frida_unref(processlist.CPtr())
		processlist = nil
	}()
	var r []*FridaProcess
	n := frida_process_list_size(processlist)
	for i := 0; i < n; i++ {
		process := frida_process_list_get(processlist, i)
		process.Fill()
		r = append(r, process)
	}
	return r, nil
}

func (this *FridaDevice) Get_Frontmost_application_sync() (*FridaApplication, error) {
	var gerr *GError
	a := frida_device_get_frontmost_application_sync(this, nil, &gerr)
	if gerr != nil {
		return nil, errors.New(gerr.Message())
	}

	if a.CPtr() == 0 {
		return nil, errors.New("没有app在前端运行")
	}
	a.Fill()
	return a, nil
}

func (this *FridaDevice) EnumApplication() ([]*FridaApplication, error) {
	var gerr *GError
	applicationlist := frida_device_enumerate_applications_sync(this, nil, &gerr)
	if gerr != nil {
		return nil, errors.New(gerr.Message())
	}
	defer func() {
		frida_unref(applicationlist.CPtr())
		applicationlist = nil
	}()
	var r []*FridaApplication
	n := frida_application_list_size(applicationlist)
	for i := 0; i < n; i++ {
		app := frida_application_list_get(applicationlist, i)
		app.Fill()
		r = append(r, app)
	}
	return r, nil
}

func (this *FridaDevice) Resume(_pid int) error {
	var gerr *GError
	frida_device_resume_sync(this, _pid, nil, &gerr)
	if gerr != nil {
		return errors.New(gerr.Message())
	}
	return nil
}
func (this *FridaDevice) Attach(_pid int) (*FridaSession, error) {
	var gerr *GError
	session := frida_device_attach_sync(this, _pid, nil, &gerr)
	if gerr != nil {
		return nil, errors.New(gerr.Message())
	}
	return session, nil
}

func (this *FridaDevice) Spawn(_cmd string) (int, error) {
	options := FridaSpawnOptions{}
	var gerr *GError

	pid := frida_device_spawn_sync(this, _cmd, &options, nil, &gerr)
	if gerr != nil {
		return 0, errors.New(gerr.Message())
	}
	return pid, nil
}

//需要有/usr/bin/launchapp 支持,ios
func (this *FridaDevice) Launchapp(pk string,label string,_args []string,timeout time.Duration)(int, error){
	newarg:=make([]string,0)
	newarg=append(newarg,pk)
	newarg=append(newarg,_args...)
	exec:="/usr/bin/launchapp"
	pid,err:=this.Spawn_args(exec,newarg)
	if err != nil {
		return 0,err
	}
	err=this.Resume(pid)
	if err != nil {
		return 0,err
	}
	t:=time.Now()
	for{
		dp,err:=this.Get_Process_with_name(label,time.Second*1)
		if err!=nil{
			if err.Error()!="Process not found"&&strings.Index(err.Error(),"Unable to find process")==-1{
				return 0,err
			}
		}else{
			return dp.Pid,nil
		}
		if time.Since(t)>timeout{
			return 0,errors.New("找不到这个pid，超时")
		}
		time.Sleep(time.Second*1)
	}
	return 0,errors.New("致命错误")
}

//启动一个停止的app
func (this *FridaDevice) Spawn_args(_path string,_args []string) (int, error) {
	options := frida_spawn_options_new()
	zarg:=make([]string,0)
	zarg=append(zarg,_path)
	zarg=append(zarg,_args...)
	options.Set_argv(zarg)
	var gerr *GError

	pid := frida_device_spawn_sync(this, _path, options, nil, &gerr)
	if gerr != nil {
		return 0, errors.New(gerr.Message())
	}
	return pid, nil
}
