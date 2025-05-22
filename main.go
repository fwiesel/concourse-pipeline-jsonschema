package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/concourse/concourse/atc"
	"github.com/invopop/jsonschema"
)

func schemaRef(schema *jsonschema.Schema) *jsonschema.Schema {
	return &jsonschema.Schema{Ref: schema.Ref}
}

func stepSchema(schema *jsonschema.Schema) {
	stepDef := schema.Definitions["Step"]

	stepVisitorType := reflect.TypeOf(struct{atc.StepVisitor}{})
	for i := range stepVisitorType.NumMethod() {
		method := stepVisitorType.Method(i)
		argument := method.Type.In(1).Elem()
		stepSchema := jsonschema.ReflectFromType(argument)
		stepDef.OneOf = append(stepDef.OneOf, schemaRef(stepSchema))
		for subName, subDef := range stepSchema.Definitions {
			if _, present := schema.Definitions[subName]; !present {
				schema.Definitions[subName] = subDef
			}
		}
	}
	stepDef.Required = nil
	stepDef.Properties = nil
	stepDef.Type = ""
	stepDef.AdditionalProperties = nil
}

func main() {
	var pipelineConfig atc.Config
	schema := jsonschema.Reflect(&pipelineConfig)
	stepSchema(schema)
	schema.AdditionalProperties = jsonschema.TrueSchema
	schema.Definitions["CheckEvery"].Type = "string"
	reflected, err := schema.MarshalJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, string(reflected))
}
