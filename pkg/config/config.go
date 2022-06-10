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

func Load(version, commit string) (Config, error) {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false,
		"show version")

	var conf Config
	flag.StringVar(&conf.FilterFile, "filter-file", "",
		"filename written in filter rules")
	flag.BoolVar(&conf.IsDebug, "is-debug", false,
		"if this is true, output debug message to stdout.")
	flag.StringVar(&conf.Endpoint, "etcd-endpoint", "http://localhost:2379/",
		"etcd endpoint URL")
	flag.StringVar(&conf.Username, "etcd-username", "",
		"etcd username (allow empty)")
	flag.StringVar(&conf.Password, "etcd-password", "",
		"etcd password (allow empty)")
	flag.StringVar(&conf.CertificateFile, "etcd-cert-file", "",
		"certification filename for etcd (allow empty)")
	flag.StringVar(&conf.CertificateKeyFile, "etcd-key-file", "",
		"certification-key filename for etcd (allow empty)")
	flag.StringVar(&conf.BasePath, "etcd-base-path", "",
		"etcd base path written by external-dns & read by CoreDNS")
	flag.StringVar(&conf.GcpCredentialFile, "gcp-credential", "",
		"GCP credential filename")
	flag.StringVar(&conf.GcpProject, "gcp-project", "",
		"GCP Project belonged Cloud DNS")
	flag.StringVar(&conf.GcpDnsManagedZone, "gcp-dns-managed-zone", "",
		"Managed Zone name of Cloud DNS")

	flag.VisitAll(func(f *flag.Flag) {
		if s := os.Getenv(strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))); s != "" {
			_ = f.Value.Set(s)
		}
	})
	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s (Commit: %s)\n", version, commit)
		return Config{}, fmt.Errorf("")
	}

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
