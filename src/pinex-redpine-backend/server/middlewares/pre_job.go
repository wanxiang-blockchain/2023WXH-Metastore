package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"

	"github.com/MetaDataLab/web3-console-backend/internal/api"
	"github.com/MetaDataLab/web3-console-backend/internal/errors"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func PreJobMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		var err error
		requestDump, err := httputil.DumpRequest(c.Request, true)
		if err != nil {
			res := api.MakeErrorResponse(errors.ParamsJudgeError("dump request failed"))
			resJson, _ := json.Marshal(res)
			log.Infof("Send response---> client %s, %s", c.ClientIP(), string(resJson))
			c.Abort()
			c.JSON(http.StatusBadRequest, res)
			return
		}

		log.Infof("--->Receive request, client %s, %s\n", c.ClientIP(), string(requestDump))

		params := make(map[string]interface{})

		contentType := c.Request.Header.Get("Content-Type")

		if len(contentType) > 0 && contentType != "application/json" {
			res := api.MakeErrorResponse(errors.ParamsJudgeError("Content-Type must be application/json"))
			resJson, _ := json.Marshal(res)
			log.Infof("Send response---> client %s, %s", c.ClientIP(), string(resJson))
			c.Abort()
			c.JSON(http.StatusBadRequest, res)
			return
		}

		err = c.BindJSON(&params)
		if err != nil {
			res := api.MakeErrorResponse(errors.ParamsJudgeError(err.Error()))
			resJson, _ := json.Marshal(res)
			log.Infof("Send response---> client %s, %s", c.ClientIP(), string(resJson))
			c.Abort()
			c.JSON(http.StatusBadRequest, res)
			return
		}
		formatJsonParams(&params)

		log.Infof("request data: %v", params)

		action, myErr := actionCheck(&params)
		if myErr.Code() != 0 {
			res := api.MakeErrorResponse(myErr)
			resJson, _ := json.Marshal(res)
			log.Infof("Send response---> client %s, %s", c.ClientIP(), string(resJson))
			c.Abort()
			c.JSON(http.StatusBadRequest, res)
			return
		}

		c.Set("action", action)
		c.Set("params", &params)

		// 继续处理请求
		c.Next()
	}
}

func actionCheck(params *map[string]interface{}) (string, errors.Error) {
	action := (*params)["Action"]
	if action == nil {
		return "", errors.MissAction("")
	}
	if !api.Exist(action.(string)) {
		return "", errors.ParamsError("Action")
	}
	return action.(string), errors.New(0, "").(errors.Error)
}

func formatJsonParams(params *map[string]interface{}) error {
	if (*params)["RequestUUID"] == nil {
		(*params)["RequestUUID"] = uuid.NewString()
	}

	return nil
}
