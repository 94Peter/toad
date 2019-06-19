package db

import (
	"fmt"

	influx "github.com/influxdata/influxdb/client/v2"
)

type influxDB struct {
	host string
	db   string

	client influx.Client
}

func (ib *influxDB) getClient() (influx.Client, error) {
	if ib.client != nil {
		return ib.client, nil
	}
	c, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr: fmt.Sprintf("http://%s", ib.host),
	})

	if err != nil {
		return nil, err
	}

	ib.client = c
	return ib.client, nil
}

func (ib *influxDB) Query(cmd string) (res []influx.Result, err error) {
	c, err := ib.getClient()
	if err != nil {
		return nil, err
	}
	q := influx.Query{
		Command:  cmd,
		Database: ib.db,
	}
	response, err := c.Query(q)
	if err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return res, response.Error()
	}
	return response.Results, nil
}

func (ib *influxDB) Save(points ...InterPoint) error {
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  ib.db,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	for _, p := range points {
		if p != nil {
			pt, err := influx.NewPoint(p.GetMeasurement(), p.GetTags(), p.GetFields(), p.GetTime())
			if err != nil {
				return err
			}
			bp.AddPoint(pt)
		}
	}
	c, err := ib.getClient()
	if err != nil {
		return err
	}
	defer c.Close()
	return c.Write(bp)
}

func (ib *influxDB) Close() error {
	if ib.client == nil {
		return nil
	}
	return ib.client.Close()
}

func (ib *influxDB) CreateDB() error {
	_, err := ib.Query(fmt.Sprintf("CREATE DATABASE %s", ib.db))
	if err != nil {
		return err
	}

	return nil
}

func (ib *influxDB) IsDBExist() bool {
	result, err := ib.Query(fmt.Sprintf("SHOW DATABASES"))

	if err != nil {
		return false
	}
	for _, r := range result {
		for _, s := range r.Series {
			for _, v := range s.Values {
				if v[0].(string) == ib.db {
					return true
				}
			}
		}
	}
	return false
}
