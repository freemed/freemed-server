package model

import (
	"fmt"
	"reflect"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/remitt-server/model"
)

// EmrModuleGetRecord retrieves a single patient EMR module record
func EmrModuleGetRecord(patient int64, module string, id int64) (common.EmrModule, error) {
	// Resolve
	mod, ok := common.EmrModuleMap[module]
	if !ok {
		return nil, fmt.Errorf("unable to resolve EMR module type %s", module)
	}

	obj := mod.Type

	// Retrieve what we need
	_, err := model.DbMap.Get(&obj, id)
	if err != nil {
		return nil, err
	}

	// Check to make sure that this belongs to patient
	e := reflect.ValueOf(&obj).Elem()

	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		//varType := e.Type().Field(i).Type
		varValue := e.Field(i).Interface()
		// Assume "Patient" field if nothing is defined
		if (varName == mod.PatientField || varName == "Patient") && varValue == patient {
			return obj, nil
		}
		//fmt.Printf("%v %v %v\n", varName,varType,varValue)
	}

	return nil, fmt.Errorf("could not identify EMR segment with specific patient")
}
