package gofrida

import (
	"errors"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"time"
)

type FridaTask struct {
	Pk string
	Label string
	Args []string
	StartTimeout time.Duration
	ScriptContext string
	Launchmode int
	Pid int
	D* FridaDevice
	Session *FridaSession
	Script *FridaScript
	RError error
	R chan bool
	T time.Duration //运行时间
	ST time.Time  //开始时间
	CB FridaTaskCB
	State int32
}
type FridaTaskCB func(task *FridaTask, message map[string]interface{}, data []byte, userdata uintptr)
const(
	LAUNCH_LAUNCH=iota
	LAUNCH_SPAWN
)

func NewFridaTask(d *FridaDevice,pk string,label string,args[]string,starttimeout time.Duration,script string,launchmode int,cb FridaTaskCB)(*FridaTask){
	this:=&FridaTask{}
	this.D=d
	this.Pk=pk
	this.Label=label
	this.Args=args
	this.StartTimeout=starttimeout
	this.ScriptContext=script
	this.Launchmode=launchmode
	this.CB=cb
	return this
}

func (this *FridaTask)Free()error{
	if this.Script!=nil{
		err:=this.Script.UnLoad()
		if err!=nil{
			return err
		}
		this.Script=nil
	}
	if this.Session!=nil{
		this.Session.Detach()
		this.Session=nil
	}
	return nil
}
func (this *FridaTask) Done(){
	if atomic.LoadInt32(&this.State)==0{
		atomic.StoreInt32(&this.State,1)
		this.R<-true
	}
}
func (this *FridaTask)Set_error(err error){
	this.RError=err
	this.Done()
}

func (this *FridaTask)Message_isjserror(message map[string]interface{})bool{
	if message["columnNumber"] != nil {
		return true
	}
	return false
}
func (this *FridaTask)Message_isdone(message map[string]interface{})bool{
	if this.Message_isjserror(message)==true{
		return false
	}
	if message["payload"] != nil {
		switch message["payload"].(type) {
		case map[string]interface{}:
			payload := message["payload"].(map[string]interface{})
			if payload["done"] != nil {
				return true
			}
		default:
		}
	}
	return false
}
func (this *FridaTask)Message_iserror(message map[string]interface{})bool{
	if this.Message_isjserror(message)==true{
		return false
	}
	if message["payload"] != nil {
		switch message["payload"].(type) {
		case map[string]interface{}:
			payload := message["payload"].(map[string]interface{})
			if payload["error"]!=nil{
				return true
			}
		default:
		}
	}
	return false
}
func (this *FridaTask)Message_payload(message map[string]interface{})bool{
	if this.Message_isjserror(message)==true{
		return false
	}
	if message["payload"] != nil {
		switch message["payload"].(type) {
		case map[string]interface{}:
			return true
		default:
		}
	}
	return false
}


func (this *FridaTask) on_message(_script *FridaScript, _message map[string]interface{}, _data []byte, _userdata uintptr){
	if atomic.LoadInt32(&this.State)==1{
		return
	}
	if this.Message_isjserror(_message) {
		this.Set_error(errors.New(_message["stack"].(string)))
		return
	}
	if this.Message_isdone(_message){
		this.Done()
		return
	}
	if this.Message_iserror(_message){
		var payload=_message["payload"].(map[string]interface{})
		this.Set_error(errors.New(payload["error"].(string)))
		return
	}
	if this.CB!=nil{
		this.CB(this,_message,_data,_userdata)
	}
}

func (this*FridaTask) Run()error{
	this.R=make(chan bool)
	this.ST=time.Now()
	defer func() {
		this.T=time.Since(this.ST)
	}()
	var err error
	if this.Launchmode==LAUNCH_LAUNCH{
		this.Pid,err=this.D.Launchapp(this.Pk,this.Label,this.Args,this.StartTimeout)
		if err!=nil{
			return err
		}
		this.Session,err=this.D.Attach(this.Pid)
		if err!=nil{
			return err
		}
		this.Script, err = this.Session.Create_Script_with_name_script("agent.js",this.ScriptContext)
		if err != nil {
			return err
		}
		this.Script.On("message",this.on_message)
		err = this.Script.Load()
		if err != nil {
			return err
		}
	}else if(this.Launchmode==LAUNCH_SPAWN){
		this.Pid,err=this.D.Spawn_args(this.Pk,this.Args)
		if err!=nil{
			return err
		}
		this.Session,err=this.D.Attach(this.Pid)
		if err!=nil{
			return err
		}
		this.Script, err = this.Session.Create_Script_with_name_script("agent.js",this.ScriptContext)
		if err != nil {
			return err
		}
		this.Script.On("message",this.on_message)
		err = this.Script.Load()
		if err != nil {
			return err
		}
		err= this.D.Resume(this.Pid)
		if err!=nil{
			return err
		}
	}else{
		return errors.New("不支持这个方式")
	}
	go func() {
		for {
			if atomic.LoadInt32(&this.State)!=0{
				break
			}
			_, err := this.D.Get_Process_with_name(this.Label,time.Second*1)
			if err != nil {
				this.Set_error(err)
				break
			}
			time.Sleep(time.Second*1)
		}
	}()
	<-this.R
	if this.RError!=nil{
		return this.RError
	}
	return nil
}

//解析script从文件开始
func ParseScriptFile(_script string,replacearr [][]string)(string,error){
	bt,err:=ioutil.ReadFile(_script)
	if err!=nil{
		return "",err
	}
	bts:=string(bt)
	for _,it:=range replacearr{
		bts=strings.ReplaceAll(bts,it[0],it[1])
	}
	return bts,nil
}