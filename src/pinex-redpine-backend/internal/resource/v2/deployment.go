package v2

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/MetaDataLab/web3-console-backend/internal/db"
	"github.com/MetaDataLab/web3-console-backend/internal/models/dao"
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/log"
	"github.com/google/uuid"
)

type PriceTag struct {
	InstancePrice float64
	NetworkPrice  float64
}

// hard code db user here
const dbUser = "root"

func toArgKey(r Role, suffix string) string {
	return strings.ToLower(r.String()) + "_" + suffix
}

type resource interface {
	Create() (string, string, error)
	Role() Role
	WaitReady(id string) error
}

type orderedResource struct {
	order *dao.CreationOrder
	r     resource
}

type deploymentConfig struct {
	chainType   ChainType
	daType      DaType
	networkType NetworkType
	keyPairId   string
	requestId   string
	*commonParams
}

type opt func(cfg *deploymentConfig)

func WithName(name string) opt {
	return func(cfg *deploymentConfig) {
		cfg.nameBase = name
	}
}

func WithProjectId(pid string) opt {
	return func(cfg *deploymentConfig) {
		cfg.projectId = pid
	}
}

func WithVpcId(vid string) opt {
	return func(cfg *deploymentConfig) {
		cfg.vpcId = vid
	}
}

func WithSubnetId(sid string) opt {
	return func(cfg *deploymentConfig) {
		cfg.subnetId = sid
	}
}

func WithKeyPairId(kid string) opt {
	return func(cfg *deploymentConfig) {
		cfg.keyPairId = kid
	}
}

func WithRequesitId(rid string) opt {
	return func(cfg *deploymentConfig) {
		cfg.requestId = rid
	}
}

type Deployment struct {
	*deploymentConfig
	tmpl       *dao.DeploymentTemplate
	statusChan chan string
	requestId  string
}

func NewDeployment(
	chainType ChainType,
	daType DaType,
	networkType NetworkType,
	proverType ProverType,
	cloud CloudVendor,
	region string,
	az string,
	token AccessToken,
	requestId string,
	opts ...opt,
) *Deployment {
	cfg := &deploymentConfig{
		chainType:   chainType,
		daType:      daType,
		networkType: networkType,
		commonParams: &commonParams{
			region:     region,
			az:         az,
			token:      token,
			cloud:      cloud,
			proverType: proverType,
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.requestId == "" {
		cfg.requestId = uuid.NewString()
	}
	return &Deployment{
		deploymentConfig: cfg,
		statusChan:       make(chan string, 10),
		requestId:        requestId,
	}
}

func (b *Deployment) LoadTemplate() (*dao.DeploymentTemplate, error) {
	if b.tmpl != nil {
		return b.tmpl, nil
	}
	tmpl, err := db.GetTemplate(b.chainType, b.daType, b.networkType, b.cloud, b.proverType, b.region)
	if err != nil {
		return nil, err
	}
	b.tmpl = tmpl
	return tmpl, err
}

func (b *Deployment) LoadPrice() (map[Role]PriceTag, error) {
	res := make(map[Role]PriceTag)
	for i := range b.tmpl.Dbs {
		db := b.tmpl.Dbs[i]
		r := NewDb(b.commonParams, &db, "", b.requestId)
		price, err := r.InstancePrice()
		if err != nil {
			return nil, err
		}
		res[db.Role] = PriceTag{
			InstancePrice: price,
			NetworkPrice:  0,
		}
	}
	for i := range b.tmpl.Nodes {
		node := b.tmpl.Nodes[i]
		r := NewVm(b.commonParams, &node, b.keyPairId, "", b.requestId, nil)
		price, err := r.InstancePrice()
		if err != nil {
			return nil, err
		}
		eipPrice, err := r.NetworkPrice()
		if err != nil {
			return nil, err
		}
		res[node.Role] = PriceTag{
			InstancePrice: price,
			NetworkPrice:  eipPrice,
		}
	}
	return res, nil
}

func (b *Deployment) StatusChan() chan string {
	return b.statusChan
}

func (b *Deployment) Build() ([]Resource, error) {
	if b.tmpl == nil {
		_, err := b.LoadTemplate()
		if err != nil {
			return nil, err
		}
	}
	return b.create()
}

// create dbs before nodes because dbs cannot be assigned ips to
func (b *Deployment) create() ([]Resource, error) {
	sort.Slice(b.tmpl.CreationOrders, func(i, j int) bool {
		return b.tmpl.CreationOrders[i].Sequence < b.tmpl.CreationOrders[j].Sequence
	})
	args := make(map[string]interface{})
	dbResources := make([]orderedResource, 0)
	nodeResources := make([]orderedResource, 0)
	orders := b.tmpl.CreationOrders

	if len(b.tmpl.Nodes)+len(b.tmpl.Dbs) != len(orders) {
		return nil, errors.New("parse creation dependencies failed, length of creation order and resources not equal")
	}
	res := make([]Resource, len(orders))
	for i := range b.tmpl.Dbs {
		db := b.tmpl.Dbs[i]
		for j := range orders {
			order := orders[j]
			if db.Role == order.Role {
				pwd := randPassword(16)
				args[toArgKey(db.Role, "password")] = pwd
				args[toArgKey(db.Role, "port")] = db.Port
				args[toArgKey(db.Role, "user")] = dbUser
				dbResources = append(dbResources, orderedResource{&order, NewDb(b.commonParams, &db, pwd, b.requestId)})
				res[order.Sequence].Type = dao.DB
				res[order.Sequence].Role = order.Role
			}
		}
	}
	sort.Slice(dbResources, func(i, j int) bool {
		return dbResources[i].order.Sequence < dbResources[j].order.Sequence
	})

	err := b.createResources(dbResources, args, res)
	if err != nil {
		return nil, err
	}

	// always allocate node ips before creation
	vpcId, subnetId, ips, err := getRandomIps(b.commonParams, len(b.tmpl.Nodes))
	if err != nil {
		return nil, err
	}
	b.commonParams.vpcId = vpcId
	b.commonParams.subnetId = subnetId

	for i := range b.tmpl.Nodes {
		node := b.tmpl.Nodes[i]
		for j := range orders {
			order := orders[j]
			if node.Role == order.Role {
				ip := ips[i]
				args[toArgKey(node.Role, "ip")] = ip
				nodeResources = append(nodeResources, orderedResource{&order, NewVm(b.commonParams, &node, b.keyPairId, ip, b.requestId, args)})
				res[order.Sequence].Type = dao.VM
				res[order.Sequence].Role = order.Role
			}
		}
	}
	sort.Slice(nodeResources, func(i, j int) bool {
		return nodeResources[i].order.Sequence < nodeResources[j].order.Sequence
	})
	err = b.createResources(nodeResources, args, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (b *Deployment) createResources(rs []orderedResource, args map[string]interface{}, res []Resource) error {
	for i := range rs {
		orderedR := rs[i]
		r := orderedR.r
		order := orderedR.order
		// middle status can be dropped
		select {
		case b.statusChan <- fmt.Sprintf("Creating %s", prettyRoleString(r.Role().String())):
		default:
		}
		log.Infow("creating "+r.Role().String(), "request id", b.requestId)
		id, ip, err := r.Create()
		if err != nil {
			return err
		}
		log.Infow(r.Role().String()+" creation done", "request id", b.requestId)
		if order.ShouldBlock {
			log.Infow("waiting for "+r.Role().String()+" ready", "request id", b.requestId)
			errChan := make(chan error)
			go func() {
				defer close(errChan)
				err = r.WaitReady(id)
				if err != nil {
					errChan <- err
				}
			}()
			ticker := time.NewTicker(20)
			var i int
		wait_loop:
			for {
				select {
				case <-ticker.C:
					i += 1
					select {
					case b.statusChan <- fmt.Sprintf("Creating %s, have been waiting for %ds", prettyRoleString(r.Role().String()), i*20):
					default:
					}
				case err = <-errChan:
					if err != nil {
						return err
					}
					break wait_loop
				}
			}

			log.Infow(r.Role().String()+" ready", "request id", b.requestId)
		}
		args[toArgKey(order.Role, "ip")] = ip
		res[order.Sequence].Id = id
	}
	return nil
}
