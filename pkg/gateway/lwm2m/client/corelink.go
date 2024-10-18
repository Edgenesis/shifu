// Reference to https://datatracker.ietf.org/doc/html/rfc6690
// unused for now
package client

import (
	"fmt"
	"strings"
)

// Link represents a single resource link with its attributes in CoRE Link Format.
type Link struct {
	ResourcePath string            // ResourcePath is the path to the resource.
	Attributes   map[string]string // Attributes are the attributes of the resource link.
}

// NewLink creates a new Link with the given resource path and attributes.
func NewLink(resourcePath string, attributes map[string]string) *Link {
	return &Link{ResourcePath: resourcePath, Attributes: attributes}
}

// String formats the Link into a CoRE Link Format string.
// example: </>;rt="oma.lwm2m",</1/0>;rt="oma.lwm2m",</1/1>;rt="oma.lwm2m"
func (l *Link) String() string {
	var attrs []string
	for key, value := range l.Attributes {
		attrs = append(attrs, fmt.Sprintf(`%s="%s"`, key, value))
	}
	return fmt.Sprintf("<%s>;%s", l.ResourcePath, strings.Join(attrs, ","))
}
