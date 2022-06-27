package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gopherlearning/gophermart/internal/args"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
)

type accrualOrder struct {
	Order string `json:"order"`
	Goods []struct {
		Description string  `json:"description"`
		Price       float64 `json:"price"`
	} `json:"goods,omitempty"`
	Status  string  `json:"status,omitempty"`
	Accrual float64 `json:"accrual,omitempty"`
}

var goods = []struct {
	Description string  "json:\"description\""
	Price       float64 "json:\"price\""
}{
	{Description: "Samsung", Price: 500000.0},
}

func (s *postgresStorage) AccrualMonitor(ctx context.Context, wg *sync.WaitGroup, url string) {
	onStop := args.StartStopFunc(ctx, wg)
	defer onStop()
	httpClient := http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxConnsPerHost:     10,
			MaxIdleConnsPerHost: 10,
		},
	}
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			rows, err := s.GetConn(ctx).Query(ctx, `select array_agg(o.id::text) as ids,s.status from orders AS o JOIN order_statuses AS s ON o.id = s.order_id WHERE s.status = ANY(ARRAY[0,2,4]) AND NOT (o.id = ANY(select o.id from orders AS o JOIN order_statuses AS s ON o.id = s.order_id WHERE s.status = ANY(ARRAY[1,3]))) GROUP BY s.status`)
			if err != nil {
				s.loger.Error(err)
				continue
			}
			// забираю и освобождаю соединение с базой
			ordersMap := make(map[v1.Order_Status][]string)
			for rows.Next() {
				var orders []string
				var status int
				err = rows.Scan(&orders, &status)
				if err != nil {
					s.loger.Error(err)
					break
				}
				ordersMap[v1.Order_Status(status)] = orders
			}
			if rows.Err() != nil {
				s.loger.Error(err)
			}
			rows.Close()
			wgLocal := &sync.WaitGroup{}
			for status, orders := range ordersMap {
				switch status {
				// case v1.Order_NEW:
				// 	for _, order := range orders {
				// 		wgLocal.Add(1)
				// 		go func(order string) {
				// 			defer wgLocal.Done()
				// 			o := accrualOrder{
				// 				Order: order,
				// 				// Goods: goods,
				// 			}
				// 			obytes, err := json.Marshal(o)
				// 			if err != nil {
				// 				s.loger.Error(err)
				// 				return
				// 			}
				// 			obuf := bytes.NewReader(obytes)
				// 			resp, err := httpClient.Post(fmt.Sprintf("%s%s", url, "/api/orders"), "application/json", obuf)
				// 			if err != nil {
				// 				s.loger.Error(err)
				// 				return
				// 			}
				// 			defer resp.Body.Close()
				// 			// s.loger.Debug(resp.StatusCode)
				// 			if resp.StatusCode != http.StatusAccepted {
				// 				return
				// 			}
				// 			_, err = s.GetConn(ctx).Exec(ctx, `INSERT INTO order_statuses (status, order_id, created_at) VALUES($1, $2, $3)`, v1.Order_REGISTERED, order, time.Now())
				// 			if err != nil {
				// 				s.loger.Error(err)
				// 				return
				// 			}
				// 		}(order)
				// 	}
				case v1.Order_PROCESSING, v1.Order_REGISTERED, v1.Order_NEW:
					for _, order := range orders {
						wgLocal.Add(1)
						go func(order string) {
							defer wgLocal.Done()
							resp, err := httpClient.Get(fmt.Sprintf("%s%s%s", url, "/api/orders/", order))
							if err != nil {
								s.loger.Error(err)
								return
							}
							defer resp.Body.Close()
							if resp.StatusCode != http.StatusOK {
								return
							}
							var o accrualOrder
							obytes, err := ioutil.ReadAll(resp.Body)
							if err != nil {
								s.loger.Error(err)
								return
							}
							s.loger.Debugf("%s", string(obytes))
							err = json.Unmarshal(obytes, &o)
							if err != nil {
								s.loger.Error(err)
								return
							}
							oInt, ok := v1.Order_Status_value[o.Status]
							if !ok {
								s.loger.Error(o.Status + " does not exist")
								return
							}
							_, err = s.GetConn(ctx).Exec(ctx, `INSERT INTO order_statuses (status, order_id, created_at) VALUES($1, $2, $3)`, oInt, order, time.Now())
							if err != nil {
								s.loger.Error(err)
								return
							}
							if o.Accrual != 0 {
								_, err = s.GetConn(ctx).Exec(ctx, `UPDATE orders SET accrual = $1 WHERE id = $2`, o.Accrual, order)
								if err != nil {
									s.loger.Error(err)
									return
								}
							}
						}(order)
					}
				}
			}
			wgLocal.Wait()
		case <-ctx.Done():
			return
		}
	}
}
