package driver

import "docker-influxdb-log-driver/influxdb"

func logMessageToInfluxdb(lp *logPair, message []byte) error {
	lp.logLine.Message = string(message[:])
	// lp.logLine.Timestamp = jsonTime{time.Now()}

	// bytes, err := json.Marshal(lp.logLine)
	// if err != nil {
	// 	return err
	// }

	return influxdb.AppendToList(lp.logLine, lp.influxdb)
}

// func (t jsonTime) MarshalJSON() ([]byte, error) {
// 	str := fmt.Sprintf("\"%s\"", t.Format(time.RFC3339Nano))
// 	return []byte(str), nil
// }
