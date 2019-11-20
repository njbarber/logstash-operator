package controller

import (
	"github.com/njbarber/logstash-operator/pkg/controller/logstash"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, logstash.Add)
}
