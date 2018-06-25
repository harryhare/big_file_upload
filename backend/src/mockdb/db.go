package mockdb

import (
	"os"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"errors"
)

type S3Part struct {
	MD5 string
	ETag string
	PartNumber int64
}
type S3Object struct {
	Parts []*S3Part
	Id *string
}
var m map[string]*S3Object

func Create(key string, reserve int,id *string){
	m[key]=&S3Object{Parts:make([]*S3Part,reserve,reserve),Id:id}
	Save()
}

func Add(key string,part *S3Part)  {
	m[key].Parts[part.PartNumber-1]=part
	Save()
}

func Get(key string)*S3Object{
	return m[key]
}

func Update(key string,part *S3Part,index int){
	m[key].Parts[index]=part
	Save()
}

func Delete(key string){
	delete(m, key)
	Save()
}

func Save()error{
	file,err:=os.Create("/Users/unity/git/big_file_upload/db")
	defer file.Close()
	if(err!=nil){
		return errors.New(fmt.Sprintf("can not open db %v\n",err))
	}
	b,err:=json.Marshal(m)
	if(err!=nil){
		return errors.New(fmt.Sprintf("can not marshal db %v\n",err))
	}
	file.Write(b)
	return  nil
}

func Load()error{
	file,err:=os.Open("/Users/unity/git/big_file_upload/db")
	defer file.Close()
	if(err!=nil){
		//return errors.New(fmt.Sprintf("can not open db %v\n",err))
		m =map[string]*S3Object{}
	}

	b,err:= ioutil.ReadAll(file)
	if(err!=nil){
		return errors.New(fmt.Sprintf("can not read db %v\n",err))
	}
	err=json.Unmarshal(b,&m)
	if(err!=nil){
		fmt.Printf("can not Unnmarshal db %v\n",err)
	}
	return nil
}