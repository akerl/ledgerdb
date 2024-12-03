package utils

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/akerl/ledgerdb/config"
)

// Point is a ledger datapoint in Influx format
type Point struct {
	Time    time.Time
	Account string
	Payee   string
	Field   string
	Value   float64
}

// WriteInflux loads transactions into Influx
func WriteInflux(c config.Config, t []Transaction) error {
	client := influxdb2.NewClient(c.InfluxURL, c.InfluxToken)
	defer client.Close()

	existing, err := getExisting(client)
	if err != nil {
		return err
	}

	actual := getActual(t)

	writeAPI := client.WriteAPIBlocking(c.InfluxOrg, c.InfluxBucket)

	for _, item := range t {
		logger.DebugMsgf("writing transaction from %s", item.Time.Format(time.DateOnly))
		p := influxdb2.NewPoint(
			"transaction",
			map[string]string{
				"account": item.Account,
				"payee":   item.Payee,
			},
			map[string]interface{}{
				"amount": item.Amount,
				"total":  item.Total,
			},
			item.Time,
		)

		err := writeAPI.WritePoint(context.Background(), p)
		if err != nil {
			return err
		}
	}

	return nil
}

func getActual(t []Transaction) map[Point]bool {
	p := map[Point]bool{}

	for _, item := range t {
		for _, point := range t.ToPoints() {
			p[point] = true
		}
	}

	return p
}

func getExisting(c config.Config, client influxdb2.Client) (map[Point]bool, error) {
	queryAPI := client.QueryAPI(c.InfluxOrg)

	query := fmt.Sprintf(`from(bucket: "%s") |> range(start: 0, stop: now())`, c.InfluxBucket)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return map[Point]bool{}, err
	}
	if result.Err() != nil {
		return map[Point]bool{}, fmt.Printf("query error: %s", result.Err().Error())
	}

	p := map[Point]bool{}
	for result.Next() {
		r := result.Record()
		newP := Point{
			Time:    r.Time(),
			Account: r.ValueByKey("account"),
			Payee:   r.ValueByKey("payee"),
			Field:   r.Field(),
			Value:   r.Value(),
		}
		p[newP] = true
	}
}
