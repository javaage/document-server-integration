/**
 *
 * (c) Copyright Ascensio System SIA 2021
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"github.com/ONLYOFFICE/document-server-integration/config"
	"github.com/ONLYOFFICE/document-server-integration/server"
	"github.com/ONLYOFFICE/document-server-integration/server/api"
	default_api "github.com/ONLYOFFICE/document-server-integration/server/api/default_module"
	bootstrapper "github.com/ONLYOFFICE/document-server-integration/server/config"
	"github.com/ONLYOFFICE/document-server-integration/server/handlers"
	default_handlers "github.com/ONLYOFFICE/document-server-integration/server/handlers/default_module"
	"github.com/ONLYOFFICE/document-server-integration/server/log"
	"github.com/ONLYOFFICE/document-server-integration/server/managers"
	default_managers "github.com/ONLYOFFICE/document-server-integration/server/managers/default_module"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(api.New),
		fx.Provide(managers.New),
		fx.Provide(handlers.New),
		fx.Provide(server.New),
		fx.Invoke(bootstrapper.Initialize),
		config.ConfigurationModule,
		log.LoggingModule,
		default_handlers.DefaultHandlersModule,
		default_managers.DefaultManagersModule,
		default_api.DefaultServerEndpointsHandlerModule,
	).Run()
}