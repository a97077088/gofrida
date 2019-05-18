package gofrida

import "C"
import (
	"errors"
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
	frida_device_manager_close_sync(this)
	frida_unref(this.CPtr())
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
	devices := frida_device_manager_enumerate_devices_sync(this, &gerr)
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
	return r, nil
}

//直接初始化devicemanager 好像没必要释放他
func init() {
	frida_init()
	DefaultDeviceManager = NewFridaDeviceManager()
}
