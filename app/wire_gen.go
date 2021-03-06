// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package app

import (
	"context"
)

// Injectors from wire.go:

func InitApplication(ctx context.Context) (*App, func(), error) {
	config, err := ProvideConfig()
	if err != nil {
		return nil, nil, err
	}
	db, cleanup, err := ProvidePostgres(config)
	if err != nil {
		return nil, nil, err
	}
	productRepository, err := ProvideProductRepository(db)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	productUsecase, err := ProvideProductUsecase(productRepository)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	productDelivery := ProvideProductDelivery(productUsecase)
	accountDelivery := ProvideAccountDelivery()
	handler := ProvideHttpHandler(productDelivery, accountDelivery)
	server := ProvideRestService(handler)
	app := &App{
		server: server,
	}
	return app, func() {
		cleanup()
	}, nil
}
