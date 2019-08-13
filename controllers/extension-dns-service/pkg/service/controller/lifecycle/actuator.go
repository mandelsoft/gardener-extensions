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
	"context"
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener-extensions/controllers/extension-dns-service"
	"github.com/gardener/gardener-extensions/controllers/extension-dns-service/pkg/imagevector"
	controllerconfig "github.com/gardener/gardener-extensions/controllers/extension-dns-service/pkg/service/controller/lifecycle/config"
	"github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/extension"
	"github.com/gardener/gardener-extensions/pkg/util"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/secrets"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	// ActuatorName is the name of the DNS Service actuator.
	ActuatorName = extension_dns_service.ServiceName + "-actuator"
	// SeedResourcesName is the name for resource describing the resources applied to the seed cluster.
	SeedResourcesName = extension_dns_service.ExtensionServiceName + "-seed"
	// ShootResourcesName is the name for resource describing the resources applied to the shoot cluster.
	ShootResourcesName = extension_dns_service.ExtensionServiceName + "-shoot"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator(config controllerconfig.Config) extension.Actuator {
	return &actuator{
		logger:           log.Log.WithName(ActuatorName),
		controllerConfig: config,
	}
}

type actuator struct {
	applier  kubernetes.ChartApplier
	renderer chartrenderer.Interface
	client   client.Client
	config   *rest.Config

	controllerConfig controllerconfig.Config

	logger logr.Logger
}

// Reconcile the Extension resource.
func (a *actuator) Reconcile(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()

	cluster, err := controller.GetCluster(ctx, a.client, namespace)
	if err != nil {
		return err
	}

	if !controller.IsHibernated(cluster.Shoot) {
		if err := a.createShootResources(ctx, cluster, ex.Namespace); err != nil {
			return err
		}
	}

	return a.createSeedResources(ctx, cluster.Shoot, namespace)
}

// Delete the Extension resource.
func (a *actuator) Delete(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()

	cluster, err := controller.GetCluster(ctx, a.client, namespace)
	if err != nil {
		return err
	}

	if err := a.deleteSeedResources(ctx, cluster.Shoot, namespace); err != nil {
		return err
	}

	return a.deleteShootResources(ctx, namespace)
}

// InjectConfig injects the rest config to this actuator.
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config
	applier, err := kubernetes.NewChartApplierForConfig(a.config)
	if err != nil {
		return fmt.Errorf("failed to create chart applier: %v", err)
	}
	a.applier = applier

	renderer, err := chartrenderer.NewForConfig(a.config)
	if err != nil {
		return fmt.Errorf("failed to create chart renderer: %v", err)
	}
	a.renderer = renderer
	return nil
}

// InjectClient injects the controller runtime client into the reconciler.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

func (a *actuator) shootId(namespace string) string {
	return fmt.Sprintf("%s.gardener.cloud/%s", a.controllerConfig.GardenID, namespace)
}

func (a *actuator) createSeedResources(ctx context.Context, shoot *gardenv1beta1.Shoot, namespace string) error {

	shootId := a.shootId(namespace)
	shootKubeconfig, err := a.createKubeconfig(ctx, namespace)
	if err != nil {
		return err
	}

	shootKubeconfigChecksum := util.ComputeChecksum(shootKubeconfig.Data)
	chartValues := map[string]interface{}{
		"serviceName":         extension_dns_service.ServiceName,
		"replicas":            util.GetReplicaCount(shoot, 1),
		"targetClusterSecret": shootKubeconfig.GetName(),
		"gardenId":            a.controllerConfig.GardenID,
		"shootId":             shootId,
		"seedId":              a.controllerConfig.SeedID,

		"podAnnotations": map[string]interface{}{
			"checksum/secret-kubeconfig": shootKubeconfigChecksum,
		},
	}

	chartValues, err = chart.InjectImages(chartValues, imagevector.ImageVector(), []string{extension_dns_service.ImageName})
	if err != nil {
		return fmt.Errorf("failed to find image version for %s: %v", extension_dns_service.ImageName, err)
	}

	a.logger.Info("Component is being applied", "component", extension_dns_service.ExtensionServiceName, "namespace", namespace)

	return a.createManagedResource(ctx, namespace, SeedResourcesName, "seed", a.renderer, extension_dns_service.SeedChartName, chartValues, nil)
}

func (a *actuator) deleteSeedResources(ctx context.Context, shoot *gardenv1beta1.Shoot, namespace string) error {

	a.logger.Info("Component is being deleted", "component", extension_dns_service.ExtensionServiceName, "namespace", namespace)

	err := controller.DeleteManagedResource(ctx, a.client, namespace, SeedResourcesName)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{}
	secret.SetName(extension_dns_service.SecretName)
	secret.SetNamespace(namespace)
	if err := client.IgnoreNotFound(a.client.Delete(context.TODO(), secret)); err != nil {
		return err
	}

	shootId := a.shootId(namespace)
	list := &unstructured.UnstructuredList{}
	list.SetAPIVersion("dns.gardener.cloud/v1alpha1")
	list.SetKind("DNSEntry")
	if err := a.client.List(context.TODO(), list, client.InNamespace(namespace), client.MatchingLabels(map[string]string{shootId: "true"})); err != nil {
		return nil
	}

	for i := range list.Items {
		if err2 := client.IgnoreNotFound(a.client.Delete(ctx, &list.Items[i])); err2 != nil {
			return err
		}
	}
	return nil
}

func (a *actuator) createShootResources(ctx context.Context, cluster *controller.Cluster, namespace string) error {
	renderer, err := util.NewChartRendererForShoot(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return errors.Wrap(err, "could not create chart renderer")
	}

	chartValues := map[string]interface{}{
		"userName":    extension_dns_service.UserName,
		"serviceName": extension_dns_service.ServiceName,
	}
	injectedLabels := map[string]string{controller.ShootNoCleanupLabel: "true"}

	return a.createManagedResource(ctx, namespace, ShootResourcesName, "", renderer, extension_dns_service.ShootChartName, chartValues, injectedLabels)
}

func (a *actuator) deleteShootResources(ctx context.Context, namespace string) error {
	return controller.DeleteManagedResource(ctx, a.client, namespace, ShootResourcesName)
}

func (a *actuator) createKubeconfig(ctx context.Context, namespace string) (*corev1.Secret, error) {
	certConfig := secrets.CertificateSecretConfig{
		Name:       extension_dns_service.SecretName,
		CommonName: extension_dns_service.UserName,
	}

	return util.GetOrCreateShootKubeconfig(ctx, a.client, certConfig, namespace)
}

func (a *actuator) createManagedResource(ctx context.Context, namespace, name, class string, renderer chartrenderer.Interface, chartName string, chartValues map[string]interface{}, injectedLabels map[string]string) error {
	return controller.CreateManagedResourceFromFileChart(
		ctx, a.client, namespace, name, class,
		renderer, filepath.Join(extension_dns_service.ChartsPath, chartName), chartName,
		chartValues, injectedLabels,
	)
}
