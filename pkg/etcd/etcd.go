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

type Client clientv3.Client

func New(conf config.Config) (*Client, error) {
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
	result := Client(*cli)
	return &result, nil
}

type KeyValue struct {
	Key   []byte
	Value []byte
}

func (c Client) List(ctx context.Context, prefix string) ([]KeyValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	resp, err := c.Get(ctx, prefix, opts...)
	if err != nil {
		return nil, err
	}

	result := []KeyValue{}
	for _, respKv := range resp.Kvs {
		result = append(result, KeyValue{
			Key:   respKv.Key,
			Value: respKv.Value,
		})
	}
	return result, nil
}

type CoreDnsRecord struct {
	Host        string `json:"host,omitempty"`
	TTL         int64  `json:"ttl,omitempty" default:"60"`
	TargetStrip int    `json:"targetstrip,omitempty"`
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
