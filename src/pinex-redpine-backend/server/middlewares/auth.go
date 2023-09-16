package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MetaDataLab/web3-console-backend/internal/api"
	"github.com/MetaDataLab/web3-console-backend/internal/errors"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

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

		authType := api.GetActionAuthType(action)
		requestUUID := (*params)["RequestUUID"].(string)

		switch authType {
		case api.COOKIEAUTH:
			cookieValue, err := c.Cookie("login_pinex")
			if err != nil {
				res := api.MakeErrorResponse(errors.ParamsJudgeError("login_pinex from cookie not found"))
				res.SetSession(requestUUID)
				res.SetAction(action + "Response")
				resJson, _ := json.Marshal(res)
				log.Infof("Send response---> client %s, %s", c.ClientIP(), string(resJson))
				c.Abort()
				c.JSON(http.StatusBadRequest, res)
				return
			}

			session := sessions.Default(c)
			addr := session.Get(cookieValue)
			if addr == nil {
				errStr := fmt.Sprintf("cookie %s not in backend session", cookieValue)
				res := api.MakeErrorResponse(errors.ParamsJudgeError(errStr))
				res.SetSession(requestUUID)
				res.SetAction(action + "Response")
				resJson, _ := json.Marshal(res)
				log.Infof("%s Send response---> client %s, %s", requestUUID, c.ClientIP(), string(resJson))
				c.Abort()
				c.JSON(http.StatusUnauthorized, res)
				return
			}

			//add addr to params
			(*params)["Address"] = addr.(string)
		}

		// 继续处理请求
		c.Next()
	}
}
