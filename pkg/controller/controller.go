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
	Log         logr.Logger
	Replacer    *replace.Replacer
	DnsService  *dnsv1.ResourceRecordSetsService
	Project     string
	ManagedZone string
	BasePath    string
}

func (r Reconciler) Reconcile(ctx context.Context, res models.WatchResponse) error {
	record, err := etcd.MarshalRecord(res.Value)
	if err != nil {
		return err
	}
	record.Host = r.Replacer.ReplaceRecord(record.Host)
	domain := r.getDomain(res.Key, record.TargetStrip)

	switch res.Type {
	case models.PUT:
		r.Log.Info("PUT", "key", string(res.Key), "value", string(res.Value))
		rrSets, err := r.DnsService.List(r.Project, r.ManagedZone).Do()
		if err != nil {
			return err
		}
		nameExists, recordIsSame := r.contain(rrSets.Rrsets, domain, record.Host)
		if recordIsSame {
			// pass
		} else if nameExists {
			if _, err := r.DnsService.Patch(
				r.Project, r.ManagedZone, domain, "A", &dnsv1.ResourceRecordSet{
					Kind:    "dns#resourceRecordSet",
					Name:    domain,
					Type:    "A",
					Rrdatas: []string{record.Host},
					Ttl:     record.TTL,
				}).Do(); err != nil {
				return err
			}
		} else {
			if _, err := r.DnsService.Create(r.Project, r.ManagedZone, &dnsv1.ResourceRecordSet{
				Kind:    "dns#resourceRecordSet",
				Name:    domain,
				Type:    "A",
				Rrdatas: []string{record.Host},
				Ttl:     record.TTL,
			}).Do(); err != nil {
				return err
			}
		}
	case models.DELETE:
		r.Log.Info("DELETE", "key", string(res.Key))
		if _, err := r.DnsService.Delete(r.Project, r.ManagedZone, domain, "A").Do(); err != nil {
			return err
		}
	}
	return nil
}

func (r Reconciler) getDomain(b []byte, targetStrip int) string {
	l := strings.Split(strings.TrimLeft(string(b), r.BasePath), "/")
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
