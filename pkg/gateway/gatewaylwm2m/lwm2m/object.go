package lwm2m

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
)

type ObjectAPI interface {
	Read() (interface{}, error)
	Write(interface{}) error
	Execute() error
}

type Object struct {
	Id    string
	Child map[string]Object
	ObjectAPI
}

type Resource struct {
	BaseName      string       `json:"bn"`
	ResourceArray []ObjectData `json:"e"`
}

func (r *Resource) ReadAsJSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

type ObjectData struct {
	ParameterName string  `json:"n,omitempty"`
	FloatValue    float64 `json:"v,omitempty"`
	StringValue   string  `json:"sv,omitempty"`
	// Pointer to avoid marshal false value
	BoolValue *bool `json:"bv,omitempty"`
}

func NewObject(Id string, data ObjectAPI) *Object {
	var objectAPI = &Object{
		Id:        Id,
		Child:     map[string]Object{},
		ObjectAPI: data,
	}
	return objectAPI
}

func (o Object) GetChildObject(path string) *Object {
	paths := strings.Split(path, "/")
	var obj *Object = &o
	for _, subPath := range paths {
		if subPath == "" {
			continue
		}

		if child, exists := obj.Child[subPath]; exists {
			obj = &child
		} else {
			return nil
		}
	}
	return obj
}

func (o Object) ReadAll(baseName string) (Resource, error) {
	var resource = Resource{}
	object := o.GetChildObject(baseName)

	data, err := object.readAll("")
	if err != nil {
		return resource, err
	}

	if len(baseName) != 0 && !strings.HasSuffix(baseName, "/") {
		baseName = baseName + "/"
	}

	resource.BaseName = baseName
	resource.ResourceArray = data
	return resource, nil
}

func (o *Object) readAll(basePath string) ([]ObjectData, error) {
	var objectDatas []ObjectData
	for _, v := range o.Child {
		path := filepath.Join(basePath, v.Id)
		childData, err := v.readAll(path)
		if err != nil {
			continue
		}

		objectDatas = append(objectDatas, childData...)
	}

	if len(o.Child) == 0 {
		data, err := o.Read()
		if err != nil {
			return nil, err
		}

		switch newData := data.(type) {
		case int, int32, int64, int16, int8:
			objectDatas = append(objectDatas, ObjectData{ParameterName: basePath, FloatValue: float64(newData.(int))})
		case float32, float64:
			objectDatas = append(objectDatas, ObjectData{ParameterName: basePath, FloatValue: newData.(float64)})
		case string:
			objectDatas = append(objectDatas, ObjectData{ParameterName: basePath, StringValue: newData})
		case bool:
			objectDatas = append(objectDatas, ObjectData{ParameterName: basePath, BoolValue: &newData})
		default:
		}
	}

	return objectDatas, nil
}

func (o *Object) AddObject(path string, childObject ObjectAPI) {
	paths := strings.Split(path, "/")
	pathEnd := len(paths) - 1
	var obj *Object = o
	for _, subPath := range paths[:pathEnd] {
		if subPath == "" {
			continue
		}

		if child, exists := obj.Child[subPath]; exists {
			obj = &child
		} else {
			newChild := Object{Id: subPath, Child: map[string]Object{}}
			obj.Child[subPath] = newChild
			obj = &newChild
		}
	}
	obj.Child[paths[pathEnd]] = *NewObject(paths[pathEnd], childObject)
}

func (o *Object) AddGroup(group Object) {
	o.Child[group.Id] = group
}

func (o *Object) GetAllChildPaths() []string {
	uniqueChildPaths := uniqueSlice(o.getChildPaths("/"))
	slices.Sort(uniqueChildPaths)
	if len(uniqueChildPaths) > 0 && uniqueChildPaths[0] == "/" {
		uniqueChildPaths = uniqueChildPaths[1:]
	}
	return uniqueChildPaths
}

func uniqueSlice(data []string) []string {
	var uniqData []string
	var dataMap = map[string]bool{}
	for _, v := range data {
		if _, exists := dataMap[v]; !exists {
			uniqData = append(uniqData, v)
			dataMap[v] = true
		}
	}

	return uniqData
}

func (o *Object) getChildPaths(basePath string) []string {
	var childPaths []string
	for k, v := range o.Child {
		childPaths = append(childPaths, v.getChildPaths(filepath.Join(basePath, k))...)
	}
	childPaths = append(childPaths, basePath)
	return childPaths
}

func (o *Object) GetCoRELinkString() string {
	childPaths := o.GetAllChildPaths()
	// Add the root path and only support json format for now
	var links []string = []string{`</>;rt="oma.lwm2m";ct="11543"`}
	for _, path := range childPaths {
		links = append(links, "<"+path+">")
	}

	return strings.Join(links, ",")
}
