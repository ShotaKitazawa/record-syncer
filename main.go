package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

const (
	defaultInterval = time.Minute * 10
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
	stream, err := etcd.Watch(ctx, conf)
	if err != nil {
		log.Error(err, "failed to initialize Etcd client")
		os.Exit(1)
	}

	// CloudDNS
	dnsService, err := dnsv1.NewService(ctx,
		option.WithCredentialsFile(conf.GcpCredentialFile),
		option.WithScopes(dnsv1.NdevClouddnsReadwriteScope))
	if err != nil {
		log.Error(err, "failed to initialize CloudDNS client")
		os.Exit(1)
	}

	r := controller.Reconciler{
		Log:         log,
		Replacer:    replacer,
		DnsService:  dnsv1.NewResourceRecordSetsService(dnsService),
		Project:     conf.GcpProject,
		ManagedZone: conf.GcpDnsManagedZone,
		BasePath:    conf.BasePath,
	}

	tick := time.Tick(defaultInterval)
	for {
		select {
		case res := <-stream:
			for _, e := range res.Events {
				err := r.Reconcile(ctx, models.WatchResponse{
					Type:  models.EventType(e.Type),
					Key:   e.Kv.Key,
					Value: e.Kv.Value,
				})
				if err != nil {
					log.Error(err, "reconcile failed")
				}
			}
		case <-tick:
			// TODO
		}
	}
}
