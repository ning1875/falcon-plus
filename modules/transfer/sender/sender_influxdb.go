package sender

import (
	"github.com/influxdata/influxdb/client/v2"
	"github.com/open-falcon/falcon-plus/modules/transfer/g"
)

func GetInfluxDbConn() (client.Client, error) {
	return client.NewHTTPClient(client.HTTPConfig{
		Addr:     g.Config().InfluxDB.Address,
		Username: g.Config().InfluxDB.UserName,
		Password: g.Config().InfluxDB.Password,
	})
}

func GetNewBatchPoints() (client.BatchPoints, error) {
	return client.NewBatchPoints(client.BatchPointsConfig{
		Database:  g.Config().InfluxDB.DBName,
		Precision: "s",
	})
}
