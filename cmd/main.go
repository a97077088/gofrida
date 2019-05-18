package main

import (
	"fmt"
	"github.com/hacdias/fileutils"
	"os"
	"path/filepath"
	"strings"
)

func Dircopy(_src,_dst string)error{
	_,err:=os.Stat(_src)
	if err!=nil{
		return err
	}
	err=os.MkdirAll(_dst,0777)
	if err!=nil{
		return err
	}
	_ = filepath.Walk(_src, func(path string, info os.FileInfo, err error) error {
		abssrc,err:=filepath.Abs(_src)
		if err!=nil{
			return err
		}
		if info.IsDir() {
			err:=os.MkdirAll(path, 0777)
			if err!=nil{
				return err
			}
		} else {
			abspath,err:=filepath.Abs(path)
			if err!=nil{
				return err
			}
			fixpath:=strings.Replace(abspath,abssrc,"",-1)
			dstpath:=fmt.Sprintf("%s\\%s",_dst,fixpath)
			_,err=os.Stat(dstpath)
			if err!=nil{
				err=fileutils.CopyFile(path,dstpath)
				if err!=nil{
					return err
				}
			}else{
				err=fileutils.CopyFile(path,dstpath)
				if err!=nil{
					return err
				}
			}
		}
		return nil
	})

	return nil
}
func getbinpath()string{
	strpaths:=os.Getenv("GOPATH")
	strpath:=strings.Split(strpaths,";")[0]
	if strpath==""{
		return ""
	}
	return fmt.Sprintf("%s\\bin",strpath)
}
func main() {
	binpath:=getbinpath()
	if binpath==""{
		panic("find gopath error")
	}
	err:=os.RemoveAll(fmt.Sprintf("%s\\frida.dll",binpath))
	if err!=nil{
		panic(err)
	}
	err=os.RemoveAll(fmt.Sprintf("%s\\vcruntime140.dll",binpath))
	if err!=nil{
		panic(err)
	}
	err=Dircopy("./bin",binpath)
	if err!=nil{
		panic(err)
	}
}
