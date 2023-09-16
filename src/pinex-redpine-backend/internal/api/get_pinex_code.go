package api

import (
	"math/rand"
	"strconv"
	"time"

	"regexp"

	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

func init() {
	Register(GET_PINEX_CODE_LABEL, NewGetPinexCodeTask, NOAUTH)
}

type GetPinexCodeRequest struct {
	BaseRequest
	Address string `mapstructure:"Address" validate:"required"`
}

type GetPinexCodeResponse struct {
	BaseResponse

	Nonce        int    `mapstructure:"Nonce"`
	CodeTemplate string `mapstructure:"CodeTemplate"`
}

type GetPinexCodeTask struct {
	Request  *GetPinexCodeRequest
	Response *GetPinexCodeResponse
}

func NewGetPinexCodeRequest(data *map[string]interface{}) (*GetPinexCodeRequest, error) {
	req := &GetPinexCodeRequest{}
	err := mapstructure.Decode(*data, &req)
	if err != nil {
		return nil, err
	}
	req.BaseRequest.RequestUUID = (*data)["RequestUUID"].(string)

	return req, nil
}

func NewGetPinexCodeResponse(sessionId string) *GetPinexCodeResponse {
	return &GetPinexCodeResponse{
		BaseResponse: BaseResponse{
			Action:      GET_PINEX_CODE_LABEL + "Response",
			RequestUUID: sessionId,
		},
	}
}

func NewGetPinexCodeTask(data *map[string]interface{}) (Task, error) {
	req, err := NewGetPinexCodeRequest(data)
	if err != nil {
		return nil, err
	}
	task := &GetPinexCodeTask{
		Request:  req,
		Response: NewGetPinexCodeResponse(req.BaseRequest.RequestUUID),
	}

	validate := validator.New()
	err = validate.Struct(task.Request)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (task *GetPinexCodeTask) Run(c *gin.Context) (Response, error) {

	var nonce int

	// save nonce redis, set TTL
	session := sessions.Default(c)
	key := MakeAddrNonceKey(task.Request.Address)
	v := session.Get(key)
	if v == nil {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		nonce = r.Intn(900000) + 100000
		session.Set(key, nonce)
		err := session.Save()
		if err != nil {
			log.Errorf("%s, save nonce to session failed, %s", task.Request.RequestUUID, err.Error())
			return nil, err
		}
	} else {
		nonce = v.(int)
	}

	reAddress := regexp.MustCompile("ADDRESS")
	reNonce := regexp.MustCompile("NONCE")

	str := codeTemplate
	str = reAddress.ReplaceAllString(str, task.Request.Address)
	str = reNonce.ReplaceAllString(str, strconv.Itoa(nonce))

	task.Response.Nonce = nonce
	task.Response.CodeTemplate = str
	return task.Response, nil
}
