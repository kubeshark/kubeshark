package oas

import (
	"github.com/chanced/openapi"
	"github.com/up9inc/mizu/shared/logger"
)

func mergePathObj(po *openapi.PathObj, other *openapi.PathObj) {
	// merge parameters
	mergeParamLists(&po.Parameters, &other.Parameters)

	// merge ops
	mergeOps(&po.Get, &other.Get)
	mergeOps(&po.Put, &other.Put)
	mergeOps(&po.Options, &other.Options)
	mergeOps(&po.Patch, &other.Patch)
	mergeOps(&po.Delete, &other.Delete)
	mergeOps(&po.Head, &other.Head)
	mergeOps(&po.Trace, &other.Trace)
	mergeOps(&po.Post, &other.Post)
}

func mergeParamLists(params **openapi.ParameterList, other **openapi.ParameterList) {
	if *other == nil {
		return
	}

	if *params == nil {
		*params = new(openapi.ParameterList)
	}

outer:
	for _, o := range **other {
		oParam, err := o.ResolveParameter(paramResolver)
		if err != nil {
			logger.Log.Warningf("Failed to resolve reference: %s", err)
			continue
		}

		for _, p := range **params {
			param, err := p.ResolveParameter(paramResolver)
			if err != nil {
				logger.Log.Warningf("Failed to resolve reference: %s", err)
				continue
			}

			if param.In == oParam.In && param.Name == oParam.Name {
				// TODO: merge examples? transfer schema pattern?
				continue outer
			}
		}

		**params = append(**params, oParam)
	}
}

func mergeOps(op **openapi.Operation, other **openapi.Operation) {
	if *other == nil {
		return
	}

	if *op == nil {
		*op = *other
	} else {
		// merge parameters
		mergeParamLists(&(*op).Parameters, &(*other).Parameters)

		// merge responses
		mergeOpResponses(&(*op).Responses, &(*other).Responses)

		// merge request body
		mergeOpReqBodies(&(*op).RequestBody, &(*other).RequestBody)

		// historical OpIDs
		appendHistoricalIds(op, other)

		// merge kpis
	}
}

func loadHistoricalIds(op *openapi.Operation) []string {
	res := make([]string, 0)
	if _, ok := op.Extensions.Extension(HistoricalIDs); ok {
		err := op.Extensions.DecodeExtension(HistoricalIDs, &res)
		if err != nil {
			logger.Log.Warningf("Failed to decode extension %s: %s", HistoricalIDs, err)
		}
	}
	return res
}

func appendHistoricalIds(op **openapi.Operation, other **openapi.Operation) {
	mIDs := loadHistoricalIds(*op)
	oIDs := loadHistoricalIds(*other)
	mIDs = append(mIDs, oIDs...)
	mIDs = append(mIDs, (*other).OperationID)

	if (*op).Extensions == nil {
		(*op).Extensions = make(openapi.Extensions)
	}

	if len(mIDs) > 1 {
		logger.Log.Debugf("")
	}

	err := (*op).Extensions.SetExtension(HistoricalIDs, mIDs)
	if err != nil {
		logger.Log.Warningf("Failed to set extension %s: %s", HistoricalIDs, err)
		return
	}
}

func mergeOpReqBodies(r *openapi.RequestBody, r2 *openapi.RequestBody) {

}

func mergeOpResponses(r *openapi.Responses, o *openapi.Responses) {

}
