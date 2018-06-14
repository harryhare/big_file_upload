package mockdb

type S3Part struct {
	MD5 string
	ETag string
}
type S3Object struct {
	Parts []*S3Part
	Id string
}
var m = map[string]*S3Object{}

func Create(key string, reserve int,id string){
	m[key]=&S3Object{Parts:make([]*S3Part,0,reserve),Id:id}
}

func Add(key string,part *S3Part)  {
	m[key].Parts=append(m[key].Parts,part)
}

func Get(key string)*S3Object{
	return m[key]
}

func Update(key string,part *S3Part,index int){
	m[key].Parts[index]=part
}

func Delete(key string){
	delete(m, key)
}