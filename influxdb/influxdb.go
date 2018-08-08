package influxdb

import (
	"encoding/json"
	"fmt"
	"time"
	"strings"

	"docker-influxdb-log-driver/commons"

	"github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
)

// Config - Configuration of DB
type Config struct {
	Server   string
	Port     int
	Database string
	Table    string
}

// Connection - structure of connection pool
type Connection struct {
	client *client.Client
	config *Config
}

type jsonTime struct {
	time.Time
}

// Connect - is main connecting function
func Connect(c *Config) *Connection {
	var err error
	var ifxdb client.Client

	// Create connect
	ifxdb, err = client.NewHTTPClient(client.HTTPConfig{
		Addr: fmt.Sprintf("http://%s:%d", c.Server, c.Port),
	})

	if err != nil {
		logrus.WithError(err).Errorf("error connecting to InfluxDB server: %q, %q", c.Server, err.Error())
	}

	// ping
	_, _, err = ifxdb.Ping(0)
	if err != nil {
		logrus.WithError(err).Errorf("Error pinging InfluxDB Cluster: %q", err.Error())
	}

	// Create database
	dbExist := false

	q := client.NewQuery("SHOW DATABASES", "", "ns")
	if response, err := ifxdb.Query(q); err == nil && response.Error() == nil {
		for _, data := range response.Results[0].Series[0].Values {
			if data[0].(string) == c.Database {
				dbExist = true
			}
		}
	} else {
		logrus.WithError(err).Errorf("Error on show databases: %q, %q", err.Error(), response.Error())
	}

	if !dbExist {
		q := client.NewQuery(fmt.Sprintf("CREATE DATABASE %s WITH DURATION 90d", c.Database), "", "")
		if response, err := ifxdb.Query(q); err != nil && response.Error() != nil {
			logrus.WithError(err).Errorf("Error on create database: %q, %q", err.Error(), response.Error())
		}
	}

	return &Connection{
		client: &ifxdb,
		config: c,
	}
}

// Disconnect - is disconnection function
func (c *Connection) Disconnect() {
	if c.client != nil {
		client := *c.client
		if err := client.Close(); err != nil {
			logrus.WithError(err).Errorf("error disconnecting to Influxdb server: %q, %q", c.config.Server, err.Error())
		}
	}
	return
}

// AppendToList - is insert data to DB
func AppendToList(logLine commons.JSONLogLine, c *Connection) error {
	var err error

	bytes, err := json.Marshal(logLine.Extra)
	if err != nil {
		return err
	}

	ifxdb := *c.client
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  c.config.Database,
		Precision: "ms",
	})

	if err != nil {
		logrus.WithError(err).Errorf("Create new row for store into table: %q", err.Error())
		return err
	}

	// Indexed atributes
	tags := map[string]string{
		"container_id":   logLine.ContainerID,
		"container_name": logLine.ContainerName,
		"image_id":       logLine.ImageID,
		"image_name":     logLine.ImageName,
		"tag":            logLine.Tag,
		"host":           logLine.Host,
	}

	// Non-indexed atributes
	fields := map[string]interface{}{
		"message":           logLine.Message,
		"container_created": logLine.ContainerCreated,
		"command":           logLine.Command,
		"extra":             string(bytes),
	}

	// Try to parse message as JSON
	var jsonMessage map[string]interface{}
	err = json.Unmarshal([]byte(logLine.Message), &jsonMessage)
	if err == nil {
		for key, val := range jsonMessage {
                        array, isArray := val.([]interface{});

			if key == "transactionId" {
				// We want transactionId to be indexed
				tags[key], _ = val.(string)
			} else if key == "message" && isArray {
                                // Message may be an array, we have to join it
                                var messages []string
                                for _, part := range array {
                                        switch part.(type) {
                                        case string:
                                                messages = append(messages, part.(string))
                                        default:
                                                str, _ := json.Marshal(part)
                                                messages = append(messages, string(str))
                                        }
                                }
                                fields[key] = strings.Join(messages, " ");
			} else {
				// Generic field
				switch val.(type) {
				case string:
					fields[key] = val.(string)
				case float64:
					fields[key] = val.(float64)
				case bool:
					fields[key] = val.(bool)
				default:
					str, _ := json.Marshal(val)
					fields[key] = string(str)
				}
			}
		}
	}

        // Insert
	point, err := client.NewPoint(
		c.config.Table,
		tags,   // Indexed
		fields, // Non-indexed
		time.Now(),
	)
	if err != nil {
		logrus.WithError(err).Errorf("Create new point: %q", err.Error())
		return err
	}

	bp.AddPoint(point)

	err = ifxdb.Write(bp)
	if err != nil {
		logrus.WithError(err).Errorf("Error on write to db: %q", err.Error())
		return err
	}

	return nil
}
