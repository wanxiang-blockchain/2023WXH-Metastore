/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/MetaDataLab/web3-console-backend/internal/cache"
	"github.com/MetaDataLab/web3-console-backend/internal/cron"
	"github.com/MetaDataLab/web3-console-backend/internal/db"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/MetaDataLab/web3-console-backend/internal/session"
	"github.com/MetaDataLab/web3-console-backend/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var configPath string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run web3-console-backend service",
	Run: func(cmd *cobra.Command, args []string) {
		err := config.InitConfig(configPath)
		if err != nil {
			fmt.Printf("load config failed: %+v", err)
			os.Exit(-1)
		}
		err = log.InitGlobalLogger(&config.GConf.LogConfig, zap.AddCallerSkip(1))
		if err != nil {
			fmt.Printf("init logger failed: %+v", err)
			os.Exit(-1)
		}
		err = db.Init()
		if err != nil {
			log.Fatalf("init db failed: %+v", err)
		}
		err = session.Init()
		if err != nil {
			log.Fatalf("init session failed: %+v", err)
		}
		cache.InitJsonCode()
		err = cache.InitGlobalCache(cache.MAP)
		if err != nil {
			log.Fatalf("init cache failed: %+v", err)
		}
		s := server.NewServer()
		s.Run()
		log.Infof("http server running on: 0.0.0.0:%d", config.GConf.Port)
		cron.RunTaskTimeoutDeamon()
		log.Infof("task timeout deamon running")
		sigCh := make(chan os.Signal)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Errorf("signal received: %s, server shutting down: %+v", sig, err)
		err = s.Stop()
		if err != nil {
			log.Fatalf("error shutting down server: %+v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	runCmd.MarkFlagRequired("config")

}
