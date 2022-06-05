# record-syncer

record-syncer syncs from etcd (used by CoreDNS) to Google CloudDNS.

<img height="300" width="900" src="https://user-images.githubusercontent.com/19530785/171987509-b7e87a4d-369e-47ef-a22b-f6425935ea2f.png">

### Motivation

Want to register pairs of home domains (managed by internal CoreDNS) and home's global address to Google CloudDNS.

### Usage

* example of execution

```bash
record-syncer --filter-file example.yml \
  --etcd-endpoint=http://localhost:2379 --etcd-base-path=/skydns \
  --gcp-credential=/tmp/credential.json --gcp-project ${GCP_PROJECT} --gcp-dns-managed-zone ${MANAGED_ZONE}
```
