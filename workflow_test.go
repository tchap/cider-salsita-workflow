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
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/cider/go-cider/cider/services/logging"
	"github.com/cider/go-cider/cider/services/pubsub"
)

func TestWorkflow(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	// Follow what go-github does!
	pivotal.ClientFactory = pivotal.MockClientFactory
	poblano.ClientFactory = poblano.MockClientFactory

	ginkgo.RunSpecs(t, "Workflow Suite")
}

var _ = ginkgo.Describe("Workflow", func() {
	var workflow *Workflow

	ginkgo.BeforeEach(func() {
		workflow = &Workflow{
			PubSub:  pubsub.NewMockClient(),
			Logging: logging.NewDummyClient(),
		}
	})

	ginkgo.AfterEach(func() {
