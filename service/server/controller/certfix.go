package controller

import (
	"net"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/server/service"
)

// PostCertFixDetect receives a list of node identifiers, scans them for TLS
// certificate risk, and returns the candidates that may need fixing.
func PostCertFixDetect(ctx *gin.Context) {
	var req struct {
		Whiches []*configure.Which `json:"whiches"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}

	candidates, err := service.DetectCandidates(req.Whiches)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}

	common.ResponseSuccess(ctx, gin.H{
		"needsCertFix":      len(candidates) > 0,
		"certFixCandidates": candidates,
	})
}

// PostCertFix starts a cert-fix job for the supplied candidates.
func PostCertFix(ctx *gin.Context) {
	var req struct {
		Candidates []service.CertFixCandidate `json:"candidates"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}

	job, err := service.StartFix(req.Candidates)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}

	common.ResponseSuccess(ctx, gin.H{
		"jobId":  job.ID,
		"status": job.Status,
	})
}

// GetCertFix returns the current state of a cert-fix job.
func GetCertFix(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		common.ResponseError(ctx, logError("missing job id"))
		return
	}

	job, err := service.GetJob(id)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}

	common.ResponseSuccess(ctx, job)
}

// DeleteCertFix cancels a running cert-fix job.
func DeleteCertFix(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		common.ResponseError(ctx, logError("missing job id"))
		return
	}

	if err := service.CancelJob(id); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}

	common.ResponseSuccess(ctx, gin.H{})
}

// certFixResponseAugment takes the standard import/subscription response data
// and adds cert-fix detection metadata. It is used by PostImport and
// PutSubscription after nodes have been persisted.
func certFixResponseAugment(ctx *gin.Context, data gin.H, whiches []*configure.Which) {
	if len(whiches) == 0 {
		common.ResponseSuccess(ctx, data)
		return
	}

	candidates, err := service.DetectCandidates(whiches)
	if err != nil {
		// Detection failure should not break the import response.
		logError(err)
		common.ResponseSuccess(ctx, data)
		return
	}

	data["needsCertFix"] = len(candidates) > 0
	data["certFixCandidates"] = candidates
	common.ResponseSuccess(ctx, data)
}

// collectImportWhiches returns the node identifier for the most recently
// appended server matching the given endpoint. It is used after a single-node
// import so the frontend can prompt for cert fixing.
func collectImportWhiches(obj serverObj.ServerObj) []*configure.Which {
	key := obj.GetProtocol() + "://" + net.JoinHostPort(obj.GetHostname(), strconv.Itoa(obj.GetPort()))
	servers := configure.GetServers()
	for i := len(servers) - 1; i >= 0; i-- {
		if servers[i].ServerObj == nil {
			continue
		}
		o := servers[i].ServerObj
		candidateKey := o.GetProtocol() + "://" + net.JoinHostPort(o.GetHostname(), strconv.Itoa(o.GetPort()))
		if candidateKey == key {
			return []*configure.Which{{TYPE: configure.ServerType, ID: i + 1}}
		}
	}
	return nil
}
