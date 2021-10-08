package api

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"

	"github.com/up9inc/mizu/shared"
)

func loadOAS() (ctx context.Context, doc *openapi3.T, router routers.Router, err error) {
	path := fmt.Sprintf("%s/%s", shared.RulePolicyPath, shared.ContractFileName)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		return
	}
	ctx = context.Background()
	loader := &openapi3.Loader{Context: ctx}
	doc, _ = loader.LoadFromFile(path)
	err = doc.Validate(ctx)
	if err != nil {
		return
	}
	router, _ = legacyrouter.NewRouter(doc)
	return
}

func validateOAS(ctx context.Context, doc *openapi3.T, router routers.Router, req *http.Request, res *http.Response) (bool, error) {
	// Find route
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		return false, err
	}

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}
	if err := openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
		return false, err
	}

	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 res.StatusCode,
		Header:                 res.Header,
	}

	body, _ := ioutil.ReadAll(res.Body)
	res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	responseValidationInput.SetBodyBytes(body)

	// Validate response.
	if err := openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
		return false, err
	}

	return true, nil
}
