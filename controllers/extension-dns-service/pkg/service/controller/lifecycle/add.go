// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lifecycle

import (
	"github.com/gardener/gardener-extensions/controllers/extension-dns-service"
	"github.com/gardener/gardener-extensions/controllers/extension-dns-service/pkg/service"
	"github.com/gardener/gardener-extensions/pkg/controller/extension"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

// ControllerName is the name o
const ControllerName = "lifecycle"
// FinalizerSuffix is the finalizer suffix for the Certificate Service controller.
const FinalizerSuffix = extension_dns_service.ExtensionServiceName


// AddToManager adds a DNS Service Lifecycle controller to the given Controller Manager.
func AddToManager(mgr manager.Manager) error {
	return extension.Add(mgr, extension.AddArgs{
		Actuator:          NewActuator(service.ServiceConfig.LifecycleConfig),
		ControllerOptions: service.ServiceConfig.ControllerOptions,
		Name:              ControllerName,
		FinalizerSuffix:   FinalizerSuffix,
		Resync:            60 * time.Minute,
		Predicates:        extension.DefaultPredicates(extension_dns_service.ExtensionType, service.ServiceConfig.IgnoreOperationAnnotation),
	})
}
