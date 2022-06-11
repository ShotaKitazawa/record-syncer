package etcd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ShotaKitazawa/record-syncer/pkg/config"
	"github.com/creasty/defaults"
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

type CoreDnsRecord struct {
	Host        string `json:"host,omitempty"`
	TTL         int64  `json:"ttl,omitempty" default:"60"`
	TargetStrip int    `json:"target_strip,omitempty"`
}

func MarshalRecord(b []byte) (*CoreDnsRecord, error) {
	record := &CoreDnsRecord{}
	if err := json.Unmarshal(b, record); err != nil {
		return nil, err
	}
	if err := defaults.Set(record); err != nil {
		return nil, err
	}
	return record, nil
}
