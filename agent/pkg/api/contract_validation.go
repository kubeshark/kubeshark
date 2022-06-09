package api

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap/api"
)

const (
	ContractNotApplicable api.ContractStatus = 0
	ContractPassed        api.ContractStatus = 1
	ContractFailed        api.ContractStatus = 2
)

func loadOAS(ctx context.Context) (doc *openapi3.T, contractContent string, router routers.Router, err error) {
	path := fmt.Sprintf("%s%s", shared.ConfigDirPath, shared.ContractFileName)
	bytesValue, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	contractContent = string(bytesValue)
	loader := &openapi3.Loader{Context: ctx}
	doc, _ = loader.LoadFromData(bytesValue)
	err = doc.Validate(ctx)
	if err != nil {
		return
	}
	router, _ = legacyrouter.NewRouter(doc)
	return
}

func validateOAS(ctx context.Context, doc *openapi3.T, router routers.Router, req *http.Request, res *http.Response) (isValid bool, reqErr error, resErr error) {
	isValid = true
	reqErr = nil
	resErr = nil

	// Find route
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		return
	}

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}
	if reqErr = openapi3filter.ValidateRequest(ctx, requestValidationInput); reqErr != nil {
		isValid = false
	}

	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 res.StatusCode,
		Header:                 res.Header,
	}

	if res.Body != nil {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		responseValidationInput.SetBodyBytes(body)
	}

	// Validate response.
	if resErr = openapi3filter.ValidateResponse(ctx, responseValidationInput); resErr != nil {
		isValid = false
	}

	return
}

func handleOAS(ctx context.Context, doc *openapi3.T, router routers.Router, req *http.Request, res *http.Response, contractContent string) (contract api.Contract) {
	contract = api.Contract{
		Content: contractContent,
		Status:  ContractNotApplicable,
	}

	isValid, reqErr, resErr := validateOAS(ctx, doc, router, req, res)
	if isValid {
		contract.Status = ContractPassed
	} else {
		contract.Status = ContractFailed
		if reqErr != nil {
			contract.RequestReason = reqErr.Error()
		} else {
			contract.RequestReason = ""
		}
		if resErr != nil {
			contract.ResponseReason = resErr.Error()
		} else {
			contract.ResponseReason = ""
		}
	}

	return
}
