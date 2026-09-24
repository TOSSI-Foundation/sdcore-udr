// Copyright (c) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package datarepository

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omec-project/openapi/v2"
	"github.com/omec-project/openapi/v2/utils"
	"github.com/omec-project/udr/logger"
	"github.com/omec-project/util/httpwrapper"
)

// writeProducerResponse serialises a producer response for handlers wired up by
// this patch. It mirrors the pattern used by the already-wired handlers (see
// api_provisioned_data_document.go) but tolerates a nil body, which some producers
// return for 204 No Content on DELETE.
func writeProducerResponse(c *gin.Context, rsp *httpwrapper.Response) {
	if rsp == nil {
		problemDetails := utils.ProblemDetailsSystemFailure("nil response from producer")
		c.JSON(http.StatusInternalServerError, problemDetails)
		return
	}

	if rsp.Body == nil {
		c.Status(rsp.Status)
		return
	}

	responseBody, err := openapi.SetBody(rsp.Body, contentTypeJSON)
	if err != nil {
		logger.DataRepoLog.Errorln(err)
		problemDetails := utils.ProblemDetailsSystemFailure(err.Error())
		c.JSON(http.StatusInternalServerError, problemDetails)
		return
	}

	c.Data(rsp.Status, contentTypeJSON, responseBody.Bytes())
}

// bindProducerJSON decodes a JSON request body into out, replying with a
// MalformedRequestSyntax ProblemDetails if it cannot be parsed.
func bindProducerJSON(c *gin.Context, out any) bool {
	if err := c.ShouldBindJSON(out); err != nil {
		logger.DataRepoLog.Errorf("malformed request body: %v", err)
		problemDetails := utils.ProblemDetailsMalformedRequestSyntax(err.Error())
		c.JSON(int(problemDetails.GetStatus()), problemDetails)
		return false
	}
	return true
}

// appDataQueryIDs collects a repeated query parameter, accepting the 3GPP name
// first and then the spellings used by existing free5GC/omec clients.
func appDataQueryIDs(c *gin.Context, names ...string) []string {
	for _, n := range names {
		if vals, ok := c.GetQueryArray(n); ok && len(vals) > 0 {
			return vals
		}
	}
	return nil
}
