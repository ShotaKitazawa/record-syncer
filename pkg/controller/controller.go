package controller

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"google.golang.org/api/dns/v1"
	dnsv1 "google.golang.org/api/dns/v1"

	"github.com/ShotaKitazawa/record-syncer/pkg/etcd"
	"github.com/ShotaKitazawa/record-syncer/pkg/models"
	"github.com/ShotaKitazawa/record-syncer/pkg/replace"
)

type Reconciler struct {
	log         logr.Logger
	replacer    *replace.Replacer
	etcdClient  *etcd.Client
	dnsService  *dnsv1.ResourceRecordSetsService
	project     string
	managedZone string
	basePath    string

	domainStorage map[string]string
}

func NewReconciler(l logr.Logger, replacer *replace.Replacer, etcdClient *etcd.Client,
	dnsService *dnsv1.ResourceRecordSetsService, project string, managedZone string, basePath string,
) *Reconciler {
	return &Reconciler{l, replacer, etcdClient, dnsService, project, managedZone, basePath, make(map[string]string)}
}

func (r Reconciler) Initialize(ctx context.Context) error {
	l, err := r.etcdClient.List(ctx, r.basePath)
	if err != nil {
		return err
	}
	for _, kv := range l {
		if err := r.Reconcile(ctx, models.WatchResponse{
			Type:  models.PUT,
			Key:   kv.Key,
			Value: kv.Value,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r Reconciler) Reconcile(ctx context.Context, res models.WatchResponse) error {
	switch res.Type {
	case models.PUT:
		record, err := etcd.MarshalRecord(res.Value)
		if err != nil {
			return err
		}
		record.Host = r.replacer.ReplaceRecord(record.Host)
		domain := r.getDomain(res.Key, record.TargetStrip)

		r.log.Info("PUT", "key", string(res.Key), "value", string(res.Value))
		rrSets, err := r.dnsService.List(r.project, r.managedZone).Do()
		if err != nil {
			return err
		}
		nameExists, recordIsSame := r.contain(rrSets.Rrsets, domain, record.Host)
		if recordIsSame {
			// pass
		} else if nameExists {
			if _, err := r.dnsService.Patch(
				r.project, r.managedZone, domain, "A", &dnsv1.ResourceRecordSet{
					Kind:    "dns#resourceRecordSet",
					Name:    domain,
					Type:    "A",
					Rrdatas: []string{record.Host},
					Ttl:     record.TTL,
				}).Do(); err != nil {
				return err
			}
		} else {
			if _, err := r.dnsService.Create(r.project, r.managedZone, &dnsv1.ResourceRecordSet{
				Kind:    "dns#resourceRecordSet",
				Name:    domain,
				Type:    "A",
				Rrdatas: []string{record.Host},
				Ttl:     record.TTL,
			}).Do(); err != nil {
				return err
			}
		}
		r.domainStorage[string(res.Key)] = domain

	case models.DELETE:
		domain, ok := r.domainStorage[string(res.Key)]
		if !ok {
			domain = string(res.Key)
		}
		r.log.Info("DELETE", "key", string(res.Key))
		if _, err := r.dnsService.Delete(r.project, r.managedZone, domain, "A").Do(); err != nil {
			return err
		}
		delete(r.domainStorage, string(res.Key))
	}
	return nil
}

func (r Reconciler) getDomain(b []byte, targetStrip int) string {
	l := strings.Split(strings.TrimLeft(string(b), r.basePath), "/")
	for i := 0; i < len(l)/2; i++ {
		l[i], l[len(l)-i-1] = l[len(l)-i-1], l[i]
	}
	l = l[targetStrip:]
	return strings.Join(append(l, ""), ".")
}

func (r Reconciler) contain(rrsets []*dns.ResourceRecordSet, name, hostname string) (bool, bool) {
	for _, rrset := range rrsets {
		if rrset.Name == name {
			for _, r := range rrset.Rrdatas {
				if r == hostname {
					return true, true
				}
			}
			return true, false
		}
	}
	return false, false
}
