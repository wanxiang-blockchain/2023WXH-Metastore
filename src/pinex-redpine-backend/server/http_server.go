package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/MetaDataLab/web3-console-backend/internal/pkg"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/MetaDataLab/web3-console-backend/internal/session"
	"github.com/MetaDataLab/web3-console-backend/server/middlewares"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/MetaDataLab/web3-console-backend/handler"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
)

type Web3FormationServer struct {
	e      *gin.Engine
	server *http.Server
	wg     *sync.WaitGroup
}

func NewServer() *Web3FormationServer {
	wg := &sync.WaitGroup{}
	r := NewRouter(wg)
	return &Web3FormationServer{
		e:  r,
		wg: wg,
	}
}

func (s *Web3FormationServer) Run() {
	s.server = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.GConf.Port),
		Handler: s.e,
	}
	go s.server.ListenAndServe()
}

func (s *Web3FormationServer) Stop() error {
	err := s.server.Shutdown(context.Background())
	// whether the err is nil, we should wait for wait group
	s.wg.Wait()
	return err
}

func NewRouter(wg *sync.WaitGroup) *gin.Engine {
	mode := strings.ToLower(config.GConf.ServerMode)
	switch mode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(ginLogger())
	r.Use(sessions.Sessions("pinex_session", session.GetStore()))
	r.Use(gin.Recovery())
	r.Use(ginWaitGroup(wg))

	// register apis here
	r.POST("/", middlewares.PreJobMiddleware(), middlewares.AuthMiddleware(), handler.Handle)

	return r
}

func ginWaitGroup(wg *sync.WaitGroup) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(pkg.SERVER_WAIT_GROUP_LABEL, wg)
	}
}

func ginLogger() gin.HandlerFunc {
	dir := config.GConf.Dir
	prefix := config.GConf.Prefix
	suffix := "http-api"
	maxAge := config.GConf.LogConfig.MaxAge
	rotationTime := config.GConf.RotationTime
	w, err := newRotateLogger(dir, prefix, suffix, maxAge, rotationTime)
	if err != nil {
		log.Fatalf("error creating http-api logger: %+v", err)
	}
	return gin.LoggerWithWriter(w)
}

func newRotateLogger(dir, prefix, suffix string, maxAge, rotationTime int) (io.Writer, error) {
	baseLogName := prefix + "." + suffix
	baseLogPath := path.Join(dir, baseLogName)
	writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(baseLogPath),                               // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(time.Duration(maxAge)*24*time.Hour),          // 文件最大保存时间
		rotatelogs.WithRotationTime(time.Duration(rotationTime)*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		return nil, err
	}

	return writer, nil
}
