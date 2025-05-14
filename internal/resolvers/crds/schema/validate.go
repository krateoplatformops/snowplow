package schema

import (
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
)

func validateCustomResource(crv *apiextensions.CustomResourceValidation, doc map[string]any) error {
	validator, _, err := validation.NewSchemaValidator(crv.OpenAPIV3Schema)
	if err != nil {
		return err
	}

	spew.Dump(doc)
	fmt.Printf("\n\n\n")
	spew.Dump(crv)

	errs := validation.ValidateCustomResource(nil, doc, validator)
	if len(errs) == 0 {
		return nil
	}

	return errors.New(errs.ToAggregate().Error())
}
