package etcd

import (
	"context"
	"crypto/tls"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ShotaKitazawa/record-syncer/pkg/config"
)

func Watch(ctx context.Context, conf config.Config) (clientv3.WatchChan, error) {
	c := clientv3.Config{
		Endpoints:   []string{conf.Endpoint},
		Username:    conf.Username,
		Password:    conf.Password,
		DialTimeout: 5 * time.Second,
	}
	if conf.CertificateFile != "" && conf.CertificateKeyFile != "" {
		certificate, err := tls.LoadX509KeyPair(conf.CertificateFile, conf.CertificateKeyFile)
		if err != nil {
			return nil, err
		}
		c.TLS = &tls.Config{Certificates: []tls.Certificate{certificate}}
	}
	cli, err := clientv3.New(c)
	if err != nil {
		return nil, err
	}
	return cli.Watch(ctx, conf.BasePath, clientv3.WithPrefix()), nil
}
