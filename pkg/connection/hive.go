package connection

import (
	"fmt"
	"tpcds_benchmark/pkg/config"

	"github.com/beltran/gohive"
)

func (cm *ConnectionManager) ConnectHive(cfg config.ConnectionConfig, engineType, database string) (*gohive.Connection, error) {

	var conn *gohive.Connection

	err := cm.retry(fmt.Sprintf("Hive(%s)", database), func() error {
		tlsConfig, err := loadTLSConfig(cm.certPath)

		if err != nil {
			return err
		}

		hiveCfg := gohive.NewConnectConfiguration()
		hiveCfg.Username = cfg.Username
		hiveCfg.Password = cfg.Password
		hiveCfg.ZookeeperNamespace = cfg.ZKNamespace
		hiveCfg.Database = database

		if len(cfg.Properties) > 0 {
			hiveCfg.HiveConfiguration = cfg.Properties
		} else if engineType != "" {
			hiveCfg.HiveConfiguration = map[string]string{
				"kyuubi.engine.type": engineType,
			}
		}

		hiveCfg.TLSConfig = tlsConfig

		c, err := gohive.ConnectZookeeper(
			cfg.ZKQuorum,
			"LDAP",
			hiveCfg,
		)

		if err != nil {
			return fmt.Errorf("ошибка соединения hive: %w", err)
		}

		conn = c
		return nil

	})

	return conn, err

}
