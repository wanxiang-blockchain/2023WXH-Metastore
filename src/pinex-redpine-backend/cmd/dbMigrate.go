/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/MetaDataLab/web3-console-backend/internal/db"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var sqlPath string

// dbMigrateCmd represents the dbMigrate command
var dbMigrateCmd = &cobra.Command{
	Use:   "db-migrate",
	Short: "create tables in the target mysql db",
	Run: func(cmd *cobra.Command, args []string) {
		err := config.InitConfig(configPath)
		if err != nil {
			fmt.Printf("load config failed: %+v", err)
			os.Exit(-1)
		}
		config.GConf.LogConfig.Dir = ""
		log.InitGlobalLogger(&config.GConf.LogConfig, zap.AddCallerSkip(1), zap.AddStacktrace(zap.DebugLevel))
		err = db.Init()
		if err != nil {
			fmt.Printf("init db failed: %+v", err)
			os.Exit(-1)
		}
		err = migrate()
		if err != nil {
			fmt.Printf("create db table failed: %+v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	rootCmd.AddCommand(dbMigrateCmd)
	dbMigrateCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	dbMigrateCmd.MarkFlagRequired("config")
}

func migrate() error {
	return db.Migrate()
}
