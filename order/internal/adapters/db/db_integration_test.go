package db

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/huseyinbabal/microservices/order/internal/application/core/domain"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type OrderDatabaseTestSuite struct {
	suite.Suite
	DataSourceUrl string
	ctx           context.Context
	cancel        context.CancelFunc
}

func (o *OrderDatabaseTestSuite) SetupTest() {
	o.ctx, o.cancel = context.WithTimeout(context.Background(), 600*time.Second)
}

func (o *OrderDatabaseTestSuite) TearDownTest() {
	o.cancel()
}

func (o *OrderDatabaseTestSuite) TearDownSuite() {
}

func (o *OrderDatabaseTestSuite) SetupSuite() {
	o.SetupTest()
	defer o.TearDownTest()
	port := "3306/tcp"
	dbURL := func(host string, port nat.Port) string {
		return fmt.Sprintf("root:meXLJ1749@tcp(%s:%s)/orders?charset=utf8mb4&parseTime=True&loc=Local", host, port.Port())
	}
	req := testcontainers.ContainerRequest{
		Image:        "docker.io/mysql:8.0.30",
		ExposedPorts: []string{port},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "meXLJ1749",
			"MYSQL_DATABASE":      "orders",
		},
		WaitingFor: wait.ForSQL(nat.Port(port), "mysql", dbURL).WithStartupTimeout(time.Second * 300),
	}
	mysqlContainer, err := testcontainers.GenericContainer(o.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal("Failed to start Mysql.", err)
	}
	endpoint, _ := mysqlContainer.Endpoint(o.ctx, "")
	o.DataSourceUrl = fmt.Sprintf("root:meXLJ1749@tcp(%s)/orders?charset=utf8mb4&parseTime=True&loc=Local", endpoint)
}

func (o *OrderDatabaseTestSuite) Test_Should_Get_Order() {
	o.SetupTest()
	defer o.TearDownTest()
	adapter, _ := NewAdapter(o.DataSourceUrl)
	order := domain.NewOrder(2, []domain.OrderItem{
		{
			ProductCode: "CAM",
			Quantity:    5,
			UnitPrice:   1.32,
		},
	})
	adapter.Save(o.ctx, &order)
	ord, _ := adapter.Get(o.ctx, order.ID)
	o.Equal(int64(2), ord.CustomerID)
}

func TestOrderDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(OrderDatabaseTestSuite))
}
