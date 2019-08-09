/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package extension_dns_service


import "path/filepath"

const ExtensionType = "dns-service"
const ServiceName = ExtensionType
const ExtensionServiceName = "extension-"+ServiceName
const SeedChartName = ServiceName+"-seed"
const ShootChartName = ServiceName+"-shoot"

// ImageName is the name of the dns controller manager.
const ImageName = "dns-controller-manager"

// UserName is the name of the user  used to connect to the target cluster.
const UserName = "dns.gardener.cloud:system:dns-service"

// SecretName is the name of the secret used to store the access data for the shoot cluster.
const SecretName = ServiceName

// ChartsPath is the path to the charts
var ChartsPath = filepath.Join("controllers", "extension-dns-service", "charts", "internal")
