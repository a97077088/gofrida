package gofrida

import "C"
import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type FridaDeviceManager struct {
	ptr uintptr
}

var DefaultDeviceManager *FridaDeviceManager

func FridaDeviceManagerFormPtr(_ptr uintptr) *FridaDeviceManager {
	return &FridaDeviceManager{_ptr}
}
func NewFridaDeviceManager() *FridaDeviceManager {
	return frida_device_manager_new()
}
func (this *FridaDeviceManager) CPtr() uintptr {
	return this.ptr
}

func (this *FridaDeviceManager) Close() {
	frida_device_manager_close_sync(this, nil, nil)
	frida_unref(this.CPtr())
}

//检测id类型是不是网络
func IsNetIp(_id string)bool{
	pr:=strings.Split(_id,".")
	if len(pr)!=4{
		return false
	}
	for _,p:=range pr{
		_,err:=strconv.Atoi(p)
		if err!=nil{
			return false
		}
	}
	return true
}
//解析地址
func ParseNetDeviceIdAddr(_id string)(string,int,error){
	var err error
	port:=27042
	pr:=strings.Split(_id,":")
	if len(pr)>1{
		port,err=strconv.Atoi(pr[1])
		if err != nil {
			return "",0, err
		}
	}
	if IsNetIp(pr[0])==false{
		return "",0,errors.New("解析ip失败")
	}
	return pr[0],port,nil
}

func (this *FridaDeviceManager) GetNetDevice_with_id_milltimeout(_id string, _milltimeout int) (*FridaDevice, error) {
	var gerr *GError
	var err error
	addr,port,err:=ParseNetDeviceIdAddr(_id)
	if err != nil {
		return nil,err
	}

	err = this.AddDeviceId(fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil,err
	}
	d := frida_device_manager_get_device_by_id_sync(this, fmt.Sprintf("tcp@%s:%d", addr,port), _milltimeout, nil, &gerr)
	if gerr != nil {
		return nil, errors.New(gerr.Message())
	}
	return d, nil
}

func (this *FridaDeviceManager) GetDevice_with_id_milltimeout(_id string, _milltimeout int) (*FridaDevice, error) {
	var gerr *GError
	d := frida_device_manager_get_device_by_id_sync(this, _id, _milltimeout, nil, &gerr)
	if gerr != nil {
		return nil, errors.New(gerr.Message())
	}
	return d, nil
}

func (this *FridaDeviceManager) EnumDevice() ([]*FridaDevice, error) {
	var gerr *GError
	devices := frida_device_manager_enumerate_devices_sync(this, nil, &gerr)
	if gerr != nil {
		return nil, errors.New(gerr.Message())
	}
	defer func() {
		frida_unref(devices.CPtr())
		devices = nil
	}()
	var r []*FridaDevice
	n := frida_device_list_size(devices)
	for i := 0; i < n; i++ {
		d := frida_device_list_get(devices, i)
		d.Type = frida_device_get_dtype(d)
		d.Name = frida_device_get_name(d)
		d.Id = frida_device_get_id(d)
		r = append(r, d)
	}
	return r, nil
}

func (this *FridaDeviceManager) GetUsbDevice() (*FridaDevice, error) {
	ds, err := this.EnumDevice()
	if err != nil {
		return nil, err
	}
	var r *FridaDevice
	for _, d := range ds {
		if d.Type == FRIDA_DEVICE_TYPE_USB {
			r = d
			break
		}
	}
	if r == nil {
		return nil, errors.New("没有找到usb设备")
	}
	return r, nil
}

func (this *FridaDeviceManager) AddDeviceId(host string)error{
	var gerr *GError
	frida_device_manager_add_remote_device_sync(this,host,nil,&gerr)
	if gerr != nil {
		return errors.New(gerr.Message())
	}
	return nil
}

//直接初始化devicemanager 好像没必要释放他
func init() {
	frida_init()
	DefaultDeviceManager = NewFridaDeviceManager()
}
