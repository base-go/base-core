package customers

import (
	"base/core/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CustomerModule struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *CustomerController
	Service    *CustomerService
}

func NewCustomerModule(db *gorm.DB, router *gin.RouterGroup) module.Module {
	service := NewCustomerService(db)
	controller := NewCustomerController(service)

	customersModule := &CustomerModule{
		DB:         db,
		Controller: controller,
		Service:    service,
	}

	customersModule.Routes(router)
	customersModule.Migrate()

	return customersModule
}

func (m *CustomerModule) Routes(router *gin.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *CustomerModule) Migrate() error {
	return m.DB.AutoMigrate(&Customer{})
}
