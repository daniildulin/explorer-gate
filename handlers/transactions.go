package handlers

import (
	"github.com/daniildulin/explorer-gate/core"
	"github.com/daniildulin/explorer-gate/errors"
	"github.com/daniildulin/explorer-gate/helpers"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/emitter"
	"net/http"
	"strings"
	"time"
)

func Index(c *gin.Context) {
	c.JSON(200, gin.H{
		"name":    "Minter Explorer Gate API",
		"version": "1.0",
	})
}

type PushTransactionRequest struct {
	Transaction string `form:"transaction" json:"transaction" binding:"required"`
}

func PushTransaction(c *gin.Context) {

	var err error
	gate, ok := c.MustGet("gate").(*core.MinterGate)
	helpers.CheckErrBool(ok)
	ee, ok := c.MustGet("emitter").(*emitter.Emitter)
	helpers.CheckErrBool(ok)

	var tx PushTransactionRequest
	if err = c.ShouldBindJSON(&tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := gate.TxPush(tx.Transaction)

	if err != nil {
		errors.SetErrorResponse(err, c)
	} else {
		select {
		case <-ee.On(strings.ToUpper(tx.Transaction), emitter.Once):
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"hash": &hash,
				},
			})
		case <-time.After(30 * time.Second):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code": 1,
					"log":  `Time out for transaction`,
				},
			})
		}
	}
}
