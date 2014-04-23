/*
   Copyright (C) 2013  Salsita s.r.o.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program. If not, see {http://www.gnu.org/licenses/}.
*/

package main

import (
	// Stdlib
	"os"
	"os/signal"
	"syscall"

	// Workflow
	"cider-salsita-workflow/poblano/v1/poblano"

	// Cider
	"github.com/cider/go-cider/cider/services/logging"
	"github.com/cider/go-cider/cider/services/pubsub"
	zlogging "github.com/cider/go-cider/cider/transports/zmq3/logging"
	zpubsub "github.com/cider/go-cider/cider/transports/zmq3/pubsub"

	// Others
	zmq "github.com/pebbe/zmq3"
)

const (
	PoblanoBaseVariableName          = "POBLANO_API_BASE_URL"
	PoblanoTokenVariableName         = "POBLANO_API_TOKEN"
	PoblanoBasicUsernameVariableName = "POBLANO_API_USERNAME"
	PoblanoBasicPasswordVariableName = "POBLANO_API_PASSWORD"
)

func main() {
	// Initialise the Logging service first.
	logger, err := logging.NewService(func() (logging.Transport, error) {
		factory := zlogging.NewTransportFactory()
		factory.MustReadConfigFromEnv("CIDER_ZMQ3_LOGGING_").MustBeFullyConfigured()
		return factory.NewTransport(os.Getenv("CIDER_ALIAS"))
	})
	if err != nil {
		panic(err)
	}
	// Make sure ZeroMQ is terminated properly.
	defer func() {
		logger.Info("Waiting for ZeroMQ context to terminate...")
		logger.Close()
		zmq.Term()
	}()

	logger.Info("Logging service initialised")

	// Panic on any error returned.
	if err := innerMain(logger); err != nil {
		panic(err)
	}
}

func innerMain(logger *logging.Service) error {
	// Read the environment.
	poblanoApiBaseURL := os.Getenv(PoblanoBaseVariableName)
	if poblanoApiBaseURL == "" {
		return logger.Criticalf("%v is not set", PoblanoBaseVariableName)
	}
	poblanoApiToken := os.Getenv(PoblanoTokenVariableName)
	if poblanoApiToken == "" {
		return logger.Criticalf("%v is not set", PoblanoTokenVariableName)
	}
	poblanoBasicUsername := os.Getenv(PoblanoBasicUsernameVariableName)
	poblanoBasicPassword := os.Getenv(PoblanoBasicPasswordVariableName)

	// Initialise PubSub service from environmental variables.
	eventBus, err := pubsub.NewService(func() (pubsub.Transport, error) {
		factory := zpubsub.NewTransportFactory()
		factory.MustReadConfigFromEnv("CIDER_ZMQ3_PUBSUB_").MustBeFullyConfigured()
		return factory.NewTransport(os.Getenv("CIDER_ALIAS"))
	})
	if err != nil {
		return logger.Critical(err)
	}
	logger.Info("PubSub service initialised")

	// Initialise the workflow.
	var poblanoApiCred *poblano.Credentials
	if poblanoBasicUsername != "" && poblanoBasicPassword != "" {
		poblanoApiCred = &poblano.Credentials{
			Username: poblanoBasicUsername,
			Password: poblanoBasicPassword,
		}
		logger.Info("Poblano API Basic authentication configured")
	}

	directory, err := poblano.NewClient(poblanoApiBaseURL, poblanoApiToken, poblanoApiCred)
	if err != nil {
		return logger.Critical(err)
	}

	workflow := &Workflow{
		directory: directory,
		eventBus:  eventBus,
		logger:    logger,
	}

	// Start catching interrupts.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Link the workflow to events of some importance to it.
	_, err = eventBus.Subscribe("github.issues", logPanic(logger, workflow.AddPtTaskFromGhIssue))
	if err != nil {
		return logger.Critical(err)
	}
	logger.Info("Hook enabled: GitHub issue created -> add Pivotal Tracker story task")

	// Block until interrupted.
	select {
	case <-signalCh:
		logger.Info("Signal received, terminating...")
		if err := eventBus.Close(); err != nil {
			return logger.Critical(err)
		}
		if err := eventBus.Wait(); err != nil {
			return logger.Critical(err)
		}
	case <-eventBus.Closed():
		// This case always signals an error.
		return logger.Critical(eventBus.Wait())
	}

	return nil
}

func logPanic(log *logging.Service, handler pubsub.EventHandler) pubsub.EventHandler {
	return func(event pubsub.Event) {
		defer func() {
			if r := recover(); r != nil {
				log.Critical(r)
			}
		}()
		handler(event)
	}
}
