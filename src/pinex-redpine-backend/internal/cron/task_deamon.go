package cron

import (
	"time"

	"github.com/MetaDataLab/web3-console-backend/internal/api"
	"github.com/MetaDataLab/web3-console-backend/internal/db"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
)

func RunTaskTimeoutDeamon() {
	go func() {
		for range time.Tick(time.Second * 30) {
			deps, err := db.ListAllDeployment()
			if err != nil {
				log.Errorf("task timeout deamon list deployments failed: %+v", err)
			}
			for _, dep := range deps {
				if dep.Status != api.FAILED && dep.Status != api.RUNNING && dep.Status != api.UNKNOWN {
					if time.Now().UTC().After(dep.UpdatedAt.UTC().Add(time.Minute)) {
						dep.Status = api.FAILED
						log.Infof("found timeout task: %d, updated time: %s, now: %s", dep.ID, dep.UpdatedAt.UTC(), time.Now().UTC())
						err := db.SaveDeployment(&dep)
						if err != nil {
							log.Errorf("task timeout set deployment timeout failed, id: %d, err: %+v", dep.ID, err)
						}
					}
				}
			}
		}
	}()
}
