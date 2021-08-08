package validation

import (
	"fmt"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	et "github.com/go-playground/validator/v10/translations/en"
)

func Validate(object interface{}) (errs []string){
	validate, trans := getValidator()
	err := validate.Struct(object)
	return translateError(err, trans)
}

func translateError(err error, trans *ut.Translator) (errs []string) {
	if err == nil {
		return nil
	}
	validatorErrs := err.(validator.ValidationErrors)
	for _, e := range validatorErrs {
		translatedErr := fmt.Errorf(e.Translate(*trans)).Error()
		errs = append(errs, translatedErr)
	}
	return errs
}

func getValidator() (*validator.Validate, *ut.Translator) {
	validate := validator.New()
	english := en.New()
	trans, _ := ut.New(english, english).GetTranslator("en")
	_ = et.RegisterDefaultTranslations(validate, trans)
	return validate, &trans
}
