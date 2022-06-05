package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	FilterFile string
	IsDebug    bool
	// for Etcd
	Endpoint           string
	Username           string
	Password           string
	CertificateFile    string
	CertificateKeyFile string
	BasePath           string
	// for CloudDNS
	GcpCredentialFile string
	GcpProject        string
	GcpDnsManagedZone string
}

func Load(version string) (Config, error) {
	conf := Config{}
	flag.StringVar(&conf.FilterFile, "filter-file", "",
		"TODO")
	flag.BoolVar(&conf.IsDebug, "is-debug", false,
		"TODO")
	flag.StringVar(&conf.Endpoint, "etcd-endpoint", "http://localhost:2379/",
		"TODO")
	flag.StringVar(&conf.Username, "etcd-username", "",
		"TODO")
	flag.StringVar(&conf.Password, "etcd-password", "",
		"TODO")
	flag.StringVar(&conf.CertificateFile, "etcd-cert-file", "",
		"TODO")
	flag.StringVar(&conf.CertificateKeyFile, "etcd-key-file", "",
		"TODO")
	flag.StringVar(&conf.BasePath, "etcd-base-path", "",
		"TODO")
	flag.StringVar(&conf.GcpCredentialFile, "gcp-credential", "",
		"TODO")
	flag.StringVar(&conf.GcpProject, "gcp-project", "",
		"TODO")
	flag.StringVar(&conf.GcpDnsManagedZone, "gcp-dns-managed-zone", "",
		"TODO")

	flag.VisitAll(func(f *flag.Flag) {
		if s := os.Getenv(strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))); s != "" {
			_ = f.Value.Set(s)
		}
	})
	flag.Parse()

	if conf.FilterFile == "" {
		return Config{}, fmt.Errorf("flag --filter-file is required")
	}
	if conf.Endpoint == "" {
		return Config{}, fmt.Errorf("flag --etcd-endpoint is required")
	}
	if conf.BasePath == "" {
		return Config{}, fmt.Errorf("flag --etcd-base-path is required")
	}
	if conf.GcpCredentialFile == "" {
		return Config{}, fmt.Errorf("flag --gcp-credential is required")
	}
	if conf.GcpProject == "" {
		return Config{}, fmt.Errorf("flag --gcp-project is required")
	}
	if conf.GcpDnsManagedZone == "" {
		return Config{}, fmt.Errorf("flag --gcp-dns-managed-zone is required")
	}

	return conf, nil
}
