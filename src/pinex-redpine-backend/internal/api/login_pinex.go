package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MetaDataLab/web3-console-backend/internal/errors"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/MetaDataLab/web3-console-backend/internal/wallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

func init() {
	Register(LOGIN_PINEX_LABEL, NewLoginPinexTask, VERIFYAUTH)
}

type LoginPinexRequest struct {
	BaseRequest
	Signature string `mapstructure:"Signature" validate:"required"`
	Address   string `mapstructure:"Address" validate:"required"`
	Nonce     int    `mapstructure:"Nonce" validate:"required"`
}

type LoginPinexResponse struct {
	BaseResponse
}

type LoginPinexTask struct {
	Request  *LoginPinexRequest
	Response *LoginPinexResponse
}

func NewLoginPinexRequest(data *map[string]interface{}) (*LoginPinexRequest, error) {
	req := &LoginPinexRequest{}
	err := mapstructure.Decode(*data, &req)
	if err != nil {
		return nil, err
	}
	req.BaseRequest.RequestUUID = (*data)["RequestUUID"].(string)

	return req, nil
}

func NewLoginPinexResponse(sessionId string) *LoginPinexResponse {
	return &LoginPinexResponse{
		BaseResponse: BaseResponse{
			Action:      LOGIN_PINEX_LABEL + "Response",
			RequestUUID: sessionId,
		},
	}
}

func NewLoginPinexTask(data *map[string]interface{}) (Task, error) {
	req, err := NewLoginPinexRequest(data)
	if err != nil {
		return nil, err
	}
	task := &LoginPinexTask{
		Request:  req,
		Response: NewLoginPinexResponse(req.BaseRequest.RequestUUID),
	}

	validate := validator.New()
	err = validate.Struct(task.Request)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (task *LoginPinexTask) Run(c *gin.Context) (Response, error) {

	// check nonce exist in session
	session := sessions.Default(c)
	key := MakeAddrNonceKey(task.Request.Address)
	v := session.Get(key)
	if v == nil {
		log.Errorf("%s, key %s has no nonce", task.Request.RequestUUID, key)
		myErr := errors.AddrHasNoNonce(task.Request.Address)
		task.Response.SetRetCode(int(myErr.Code()))
		task.Response.SetMessage(myErr.String())
		return task.Response, nil
	}
	nonce := v.(int)

	if nonce != task.Request.Nonce {
		log.Errorf("%s, nonce %s != task.Request.Nonce %s", task.Request.RequestUUID, nonce, task.Request.Nonce)
		myErr := errors.NonceInvalid(strconv.Itoa(task.Request.Nonce))
		task.Response.SetRetCode(int(myErr.Code()))
		task.Response.SetMessage(myErr.String())
		return task.Response, nil
	}

	//1 verify signature
	data := strings.Replace(codeTemplate, "ADDRESS", task.Request.Address, -1)
	data = strings.Replace(data, "NONCE", strconv.Itoa(task.Request.Nonce), -1)

	ok, err := verifySign(data, task.Request.Signature, task.Request.Address)
	if err != nil || !ok {
		if err != nil {
			log.Errorf("%s, verifySign failed, err: %s", task.Request.RequestUUID, err.Error())
		} else {
			log.Errorf("%s, verifySign failed, ok: %v", task.Request.RequestUUID, ok)
		}
		task.Response.BaseResponse.RetCode = int(errors.SignatureInvalid().Code())
		task.Response.BaseResponse.Message = errors.SignatureInvalid().Message()
		return task.Response, err
	}

	//2 generate cookie
	var cookieValue string

	retryCount := 10
	isSet := false
	for i := 0; i < retryCount; i++ {
		cookieValue = uuid.NewString()
		v := session.Get(cookieValue)
		if v == nil {
			session.Set(cookieValue, task.Request.Address)
			err := session.Save()
			if err == nil {
				isSet = true
				break
			} else {
				log.Errorf("%s, save cookie-address to session failed, %s", task.Request.RequestUUID, err.Error())
			}
		}
	}

	if !isSet {
		task.Response.BaseResponse.RetCode = int(errors.SaveSessionFailed().Code())
		task.Response.BaseResponse.Message = errors.SaveSessionFailed().Message()
		return task.Response, err
	}

	cookie := &http.Cookie{
		Name:     "login_pinex",
		Value:    cookieValue,
		Expires:  time.Now().Add(12 * time.Hour), // 设置超时时间为12小时
		HttpOnly: true,                           // 仅允许通过 HTTP 访问
		Secure:   false,                          // 仅在 HTTPS 连接中传输
	}
	http.SetCookie(c.Writer, cookie)

	return task.Response, nil
}

func verifySign(message string, signature string, addr string) (bool, error) {

	signatureBytes := common.Hex2Bytes(strings.TrimPrefix(signature, "0x"))
	addrBytes := common.Hex2Bytes(strings.TrimPrefix(addr, "0x"))

	return wallet.EthVerify(message, signatureBytes, addrBytes)
}
