package influxdb

import (
	"context"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
)

func WriteData(serviceName string, containerID string, hostName string, servicePort int, serviceIP string, serviceStatus string, serviceTags []string) {

	bucketName := os.Getenv("bucket")
	influxToken := os.Getenv("influx_token")
	orgName := os.Getenv("org_name")
	influxURL := os.Getenv("influx_url")
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient(influxURL, influxToken)
	defer client.Close()
	// Use blocking write client for writes to desired bucket

	writeAPI := client.WriteAPIBlocking(orgName, bucketName)
	// write some points

	p := influxdb2.NewPointWithMeasurement("stat").
		AddTag("service_name", serviceName).
		AddField("container_id", containerID[:12]).
		AddField("host", hostName).
		AddField("port", servicePort).
		AddField("ip", serviceIP).
		AddField("status", serviceStatus).
		AddField("tags", serviceTags).
		SetTime(time.Now())

	// write synchronously
	err := writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		panic(err)
	}

	// Ensures background processes finishes
	client.Close()
}
