package driver

import "docker-influxdb-log-driver/influxdb"

func logMessageToInfluxdb(lp *logPair, message []byte) error {
	lp.logLine.Message = string(message[:])

        // Dynamic attributes are paresed in AppendToList
	return influxdb.AppendToList(lp.logLine, lp.influxdb)
}
