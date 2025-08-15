package client

import (
	"encoding/json"
	"fmt"
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

const (
	ObjectPathDelimiter = "/"
	ObjectRoot          = "/"
	defaultCoRELinkRoot = `</>;rt="oma.lwm2m";ct="11543"`
)

type Resource struct {
	BaseName      string       `json:"bn"`
	ResourceArray []ObjectData `json:"e"`
}

func (r *Resource) ReadAsJSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

type ObjectData struct {
	ParameterName *string  `json:"n,omitempty"`
	FloatValue    *float64 `json:"v,omitempty"`
	StringValue   *string  `json:"sv,omitempty"`
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

// GetChildObject returns the child object at the given path if it exists.
func (o Object) GetChildObject(path string) *Object {
	paths := strings.Split(path, "/")
	var obj *Object = &o
	for _, subPath := range paths {
		if len(subPath) == 0 {
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

// ReadAll reads all the data from the object and its children.
func (o Object) ReadAll(baseName string) (Resource, error) {
	var resource = Resource{}
	// Get the target object
	object := o.GetChildObject(baseName)

	// read all the data from the object and its children
	data, err := object.readAll("")
	if err != nil {
		return resource, err
	}

	// add slash to the base name if it does not have it
	if len(baseName) != 0 && !strings.HasSuffix(baseName, "/") {
		baseName = baseName + "/"
	}

	resource.BaseName = baseName
	resource.ResourceArray = data
	return resource, nil
}

func (o *Object) readAll(basePath string) ([]ObjectData, error) {
	var objectDataList []ObjectData
	for _, v := range o.Child {
		path := filepath.Join(basePath, v.Id)
		childData, err := v.readAll(path)
		if err != nil {
			continue
		}

		objectDataList = append(objectDataList, childData...)
	}

	if len(o.Child) == 0 {
		data, err := o.Read()
		if err != nil {
			return nil, err
		}

		switch newData := data.(type) {
		case int, int32, int64, int16, int8:
			var floatValue = float64(newData.(int))
			objectDataList = append(objectDataList, ObjectData{ParameterName: &basePath, FloatValue: &floatValue})
		case float32, float64:
			var floatValue = newData.(float64)
			objectDataList = append(objectDataList, ObjectData{ParameterName: &basePath, FloatValue: &floatValue})
		case string:
			objectDataList = append(objectDataList, ObjectData{ParameterName: &basePath, StringValue: &newData})
		case bool:
			objectDataList = append(objectDataList, ObjectData{ParameterName: &basePath, BoolValue: &newData})
		default:
			// default to string
			var stringValue = fmt.Sprintf("%v", data)
			objectDataList = append(objectDataList, ObjectData{ParameterName: &basePath, StringValue: &stringValue})
		}
	}

	return objectDataList, nil
}

// AddObject adds a child Object to target path
// input: path is the path of the child object. example: /1/0
// input: childObject is the child object to be added
func (o *Object) AddObject(path string, childObject ObjectAPI) {
	paths := strings.Split(path, "/")
	pathEnd := len(paths) - 1
	var obj *Object = o
	// iterate through the path and set the object to the last path
	for _, subPath := range paths[:pathEnd] {
		if len(subPath) == 0 {
			continue
		}

		if child, exists := obj.Child[subPath]; exists {
			obj = &child
		} else {
			// create a new child object if it does not exist
			newChild := Object{Id: subPath, Child: map[string]Object{}}
			obj.Child[subPath] = newChild
			obj = &newChild
		}
	}
	// add the child object to the last path
	obj.Child[paths[pathEnd]] = *NewObject(paths[pathEnd], childObject)
}

func (o *Object) AddGroup(group Object) {
	o.Child[group.Id] = group
}

// GetAllChildPaths returns all the child paths of the object.
// example: [/1/0, /1/1]
func (o *Object) GetAllChildPaths() []string {
	uniqueChildPaths := uniqueSlice(o.getChildPaths(ObjectRoot))
	slices.Sort(uniqueChildPaths)
	// remove base path if it exists
	if len(uniqueChildPaths) > 0 && uniqueChildPaths[0] == ObjectRoot {
		uniqueChildPaths = uniqueChildPaths[1:]
	}
	return uniqueChildPaths
}

// uniqueSlice returns a slice with unique elements.
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

// getChildPaths returns base path,all the child paths of the object.
// input basePath is the path of the object. example: / or /1 or /1/0 and so on
// output is the list of all the child paths of the object. example: [/,/1/0, /1/1]
func (o *Object) getChildPaths(basePath string) []string {
	var childPaths []string
	for k, v := range o.Child {
		childPaths = append(childPaths, v.getChildPaths(filepath.Join(basePath, k))...)
	}
	childPaths = append(childPaths, basePath)
	return childPaths
}

// GetCoRELinkString returns the CoRE Link Format string for the object and its children.
// Reference to https://datatracker.ietf.org/doc/html/rfc6690
// example: </>;rt="oma.lwm2m";ct="11543",</1/0>,</1/1>
func (o *Object) GetCoRELinkString() string {
	childPaths := o.GetAllChildPaths()
	// Add the root path and only support json format for now
	// ct=11543 is the content format for application/vnd.oma.lwm2m+json
	// reference to https://www.iana.org/assignments/core-parameters/core-parameters.xhtml#content-formats
	var links []string = []string{defaultCoRELinkRoot}
	for _, path := range childPaths {
		links = append(links, fmt.Sprintf("<%s>", path))
	}

	// return the string with comma separated links
	// example: </>;rt="oma.lwm2m";ct="11543",</1/0>,</1/1>
	return strings.Join(links, ",")
}
