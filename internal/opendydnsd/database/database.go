package database

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/opendydnsd/config"
	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:generate mockgen -source database.go -destination=../database_mock/database_mock.go -package=database_mock

// User is the mapping of an user
type User struct {
	gorm.Model

	Email    string `gorm:"unique"`
	Password string

	Aliases []Alias
}

// Alias is the mapping of a DyDNS alias
type Alias struct {
	gorm.Model

	Host   string
	Domain string
	Value  string
	UserID uint // FK
}

// Connection represent a connection to the database
// to perform CRUD
type Connection interface {
	CreateUser(email, hashedPassword string) (User, error)
	FindUser(email string) (User, error)
	FindUserAliases(userID uint) ([]Alias, error)
	FindAlias(host, domain string) (Alias, error)
	CreateAlias(alias Alias, userID uint) (Alias, error)
	DeleteAlias(host, domain string, userID uint) error
	UpdateAlias(alias Alias) (Alias, error)
}

type connection struct {
	connection *gorm.DB
}

// OpenConnection tries to open a new database connection using given config
func OpenConnection(conf config.DatabaseConfig, logger *zerolog.Logger) (Connection, error) {
	driver, err := getDriver(conf)
	if err != nil {
		return nil, err
	}

	conn, err := gorm.Open(driver, &gorm.Config{
		Logger: &zeroLogger{logger: logger},
	})
	if err != nil {
		return nil, err
	}

	// TODO remove? better?
	if err := conn.AutoMigrate(&Alias{}, &User{}); err != nil {
		return nil, err
	}

	return &connection{
		connection: conn,
	}, nil
}

func (c *connection) CreateUser(email, hashedPassword string) (User, error) {
	user := User{
		Email:    email,
		Password: hashedPassword,
	}

	result := c.connection.Create(&user)
	return user, result.Error
}

func (c *connection) FindUser(email string) (User, error) {
	var user User
	result := c.connection.Where("email = ?", email).First(&user)
	return user, result.Error
}

func (c *connection) FindUserAliases(userID uint) ([]Alias, error) {
	var aliases []Alias
	err := c.connection.Model(&User{Model: gorm.Model{ID: userID}}).Association("Aliases").Find(&aliases)
	return aliases, err
}

func (c *connection) FindAlias(host, domain string) (Alias, error) {
	var alias Alias
	result := c.connection.Where("host = ? AND domain = ?", host, domain).First(&alias)
	return alias, result.Error
}

func (c *connection) CreateAlias(alias Alias, userID uint) (Alias, error) {
	err := c.connection.Model(&User{Model: gorm.Model{ID: userID}}).Association("Aliases").Append(&alias)
	return alias, err
}

func (c *connection) DeleteAlias(host, domain string, userID uint) error {
	result := c.connection.Where("host = ? AND domain = ? AND user_id = ?", host, domain, userID).Delete(Alias{})
	return result.Error
}

func (c *connection) UpdateAlias(alias Alias) (Alias, error) {
	result := c.connection.Model(&alias).Updates(Alias{
		Domain: alias.Domain,
		Value:  alias.Value,
	})
	return alias, result.Error
}

func getDriver(conf config.DatabaseConfig) (gorm.Dialector, error) {
	switch conf.Driver {
	case "sqlite":
		return sqlite.Open(conf.DSN), nil
	default:
		return nil, fmt.Errorf("no database driver named `%s` found", conf.Driver)
	}
}
