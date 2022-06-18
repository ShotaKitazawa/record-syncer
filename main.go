package main

import (
	"context"
	"fmt"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	dnsv1 "google.golang.org/api/dns/v1"
	"google.golang.org/api/option"

	"github.com/ShotaKitazawa/record-syncer/pkg/config"
	"github.com/ShotaKitazawa/record-syncer/pkg/controller"
	"github.com/ShotaKitazawa/record-syncer/pkg/etcd"
	"github.com/ShotaKitazawa/record-syncer/pkg/logger"
	"github.com/ShotaKitazawa/record-syncer/pkg/models"
	"github.com/ShotaKitazawa/record-syncer/pkg/replace"
)

var (
	appVersion = "not_specified"
	appCommit  = "not_specified"
)

func main() {
	ctx := context.Background()
	// from env vars
	conf, err := config.Load(appVersion, appCommit)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Logger
	log, err := logger.New(conf.IsDebug)
	if err != nil {
		panic(err)
	}
	// Replacer
	replacer, err := replace.New(conf.FilterFile)
	if err != nil {
		log.Error(err, "failed to initialize replacer")
		os.Exit(1)
	}
	// Etcd
	etcdClient, err := etcd.New(conf)
	if err != nil {
		log.Error(err, "failed to initialize Etcd client")
		os.Exit(1)
	}
	stream := etcdClient.Watch(ctx, conf.BasePath, clientv3.WithPrefix())
	// CloudDNS
	dnsService, err := dnsv1.NewService(ctx,
		option.WithCredentialsFile(conf.GcpCredentialFile),
		option.WithScopes(dnsv1.NdevClouddnsReadwriteScope))
	if err != nil {
		log.Error(err, "failed to initialize CloudDNS client")
		os.Exit(1)
	}
	// Controller
	r := controller.NewReconciler(log, replacer, etcdClient,
		dnsv1.NewResourceRecordSetsService(dnsService),
		conf.GcpProject, conf.GcpDnsManagedZone, conf.BasePath,
	)
	if err := r.Resync(ctx); err != nil {
		log.Error(err, "failed to initialize Reconciler")
		os.Exit(1)
	}
	// main loop
	tick := time.Tick(time.Duration(conf.ResyncPeriodMinutes) * time.Minute)
	for {
		select {
		case res := <-stream:
			for _, e := range res.Events {
				if err := r.Reconcile(ctx, models.WatchResponse{
					Type:  models.EventType(e.Type),
					Key:   e.Kv.Key,
					Value: e.Kv.Value,
				}); err != nil {
					log.Error(err, "reconcile failed")
				}
			}
		case <-tick:
			if err := r.Resync(ctx); err != nil {
				log.Error(err, "resync failed")
			}
		}
	}
}
