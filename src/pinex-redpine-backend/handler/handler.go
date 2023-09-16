package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MetaDataLab/web3-console-backend/internal/api"
	"github.com/MetaDataLab/web3-console-backend/internal/errors"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/gin-gonic/gin"
)

func Handle(c *gin.Context) {
	var err error

	var (
		task api.Task
		res  api.Response
	)

	action := c.GetString("action")
	_params, _ := c.Get("params")
	params, ok := _params.(*map[string]interface{})
	if !ok {
		res := api.MakeErrorResponse(errors.ParamsJudgeError("params assert failed"))
		resJson, _ := json.Marshal(res)
		log.Infof("Send response---> client %s, %s", c.ClientIP(), string(resJson))
		c.Abort()
		c.JSON(http.StatusBadRequest, res)
		return
	}

	requestUUID := (*params)["RequestUUID"].(string)

	task, err = api.NewTask(action, params)
	if err != nil {
		res := api.MakeErrorResponse(errors.ParamsJudgeError(err.Error()))
		res.SetSession(requestUUID)
		res.SetAction(action + "Response")
		resJson, _ := json.Marshal(res)
		log.Infof("Send response---> client %s, %s", c.ClientIP(), string(resJson))
		c.Abort()
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res, err = task.Run(c)
	if err == nil {
		resJson, err := json.Marshal(res)
		if err == nil {
			log.Infof("Send response---> client %s, %s", c.ClientIP(), string(resJson))
			c.JSON(http.StatusOK, res)
			return
		}
	} else {
		res := api.MakeErrorResponse(errors.ActionError(err.Error()))
		res.SetSession(requestUUID)
		res.SetAction(action + "Response")
		resJson, _ := json.Marshal(res)
		c.JSON(http.StatusOK, res)
		log.Infof("Send response---> client %s, %s", c.ClientIP(), resJson)
		return
	}
}
