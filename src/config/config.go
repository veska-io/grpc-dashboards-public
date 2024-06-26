package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	DEFAULT_DEBUG = false

	DEFAULT_GRPC_PORT    = 44044
	DEFAULT_GRPC_TIMEOUT = "5s"

	DEFAULT_CLICKHOUSE_DB   = "default"
	DEFAULT_CLICKHOUSE_USER = "default"

	DEFAULT_MAXIMUM_DATA_CORRUPT = 5
)

type Config struct {
	Debug bool `koanf:"DASHBOARDS_DEBUG"`

	GrpcPort    int    `koanf:"DASHBOARDS_GRPC_PORT"`
	GrpcTimeout string `koanf:"DASHBOARDS_GRPC_TIMEOUT"`

	ClickhouseHost     string `koanf:"DASHBOARDS_STORAGE_CLICKHOUSE_HOST"`
	ClickhousePort     int    `koanf:"DASHBOARDS_STORAGE_CLICKHOUSE_PORT"`
	ClickhouseDb       string `koanf:"DASHBOARDS_STORAGE_CLICKHOUSE_DB"`
	ClickhouseUser     string `koanf:"DASHBOARDS_STORAGE_CLICKHOUSE_USER"`
	ClickhousePassword string `koanf:"DASHBOARDS_STORAGE_CLICKHOUSE_PASSWORD"`

	MaximumDataCorrupt int16 `koanf:"DASHBOARDS_MAXIMUM_DATA_CORRUPT"`
}

func MustNew() *Config {
	var c Config

	k := koanf.New(".")

	mustLoadDefaults(k)

	fileFlag := mustCheckFileFlag()
	if fileFlag != "" {
		mustLoadYamlFile(k, fileFlag)
	}

	mustLoadEnv(k)

	err := k.Unmarshal("", &c)
	if err != nil {
		panic(fmt.Errorf("error while unmarshalling config: %w", err))
	}

	mustLoadCloudRunEnv(&c)

	return &c
}

func mustLoadDefaults(k *koanf.Koanf) {
	err := k.Load(confmap.Provider(map[string]interface{}{
		"DASHBOARDS_DEBUG": DEFAULT_DEBUG,

		"DASHBOARDS_GRPC_PORT":    DEFAULT_GRPC_PORT,
		"DASHBOARDS_GRPC_TIMEOUT": DEFAULT_GRPC_TIMEOUT,

		"DASHBOARDS_STORAGE_CLICKHOUSE_DB":   DEFAULT_CLICKHOUSE_DB,
		"DASHBOARDS_STORAGE_CLICKHOUSE_USER": DEFAULT_CLICKHOUSE_USER,

		"DASHBOARDS_MAXIMUM_DATA_CORRUPT": DEFAULT_MAXIMUM_DATA_CORRUPT,
	}, "."), nil)
	if err != nil {
		panic(fmt.Errorf("error while loading config defaults: %w", err))
	}
}

func mustCheckFileFlag() string {
	var fFlag = flag.String("f", "", "Path to the configuration YAML file")

	flag.Parse()

	return *fFlag
}

func mustLoadYamlFile(k *koanf.Koanf, name string) {
	err := k.Load(file.Provider(name), yaml.Parser())
	if err != nil {
		panic(fmt.Errorf("error while loading yaml config file: %w", err))
	}
}

func mustLoadEnv(k *koanf.Koanf) {
	err := k.Load(env.Provider("DASHBOARDS_", ".", nil), nil)
	if err != nil {
		panic(fmt.Errorf("error while loading env vars: %w", err))
	}
}

func mustLoadCloudRunEnv(c *Config) {
	if port := os.Getenv("PORT"); port != "" {
		iPort, err := strconv.Atoi(port)
		if err != nil {
			panic(fmt.Errorf("error while parsing PORT env var: %w", err))
		}
		c.GrpcPort = iPort
	}
}
