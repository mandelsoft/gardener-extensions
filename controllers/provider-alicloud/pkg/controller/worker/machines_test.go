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

package worker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gardener/gardener-extensions/pkg/controller/common"
	"path/filepath"

	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud"
	api "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud"
	apiv1alpha1 "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/apis/alicloud/v1alpha1"
	. "github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/controller/worker"
	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/worker"
	mockclient "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/client"
	mockkubernetes "github.com/gardener/gardener-extensions/pkg/mock/gardener/client/kubernetes"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Machines", func() {
	var (
		ctrl         *gomock.Controller
		c            *mockclient.MockClient
		chartApplier *mockkubernetes.MockChartApplier
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		c = mockclient.NewMockClient(ctrl)
		chartApplier = mockkubernetes.NewMockChartApplier(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("workerDelegate", func() {
		workerDelegate, _ := NewWorkerDelegate(common.NewClientContext(nil, nil, nil), nil, "", nil, nil)

		Describe("#MachineClassKind", func() {
			It("should return the correct kind of the machine class", func() {
				Expect(workerDelegate.MachineClassKind()).To(Equal("AlicloudMachineClass"))
			})
		})

		Describe("#MachineClassList", func() {
			It("should return the correct type for the machine class list", func() {
				Expect(workerDelegate.MachineClassList()).To(Equal(&machinev1alpha1.AlicloudMachineClassList{}))
			})
		})

		Describe("#GenerateMachineDeployments, #DeployMachineClasses", func() {
			var (
				namespace        string
				cloudProfileName string

				alicloudAccessKeyID     string
				alicloudAccessKeySecret string
				region                  string

				machineImageName    string
				machineImageVersion string
				machineImageID      string

				instanceChargeType      string
				internetChargeType      string
				internetMaxBandwidthIn  int
				internetMaxBandwidthOut int
				spotStrategy            string

				machineType     string
				userData        []byte
				securityGroupID string
				keyName         string

				volumeType string
				volumeSize int

				namePool1           string
				minPool1            int
				maxPool1            int
				maxSurgePool1       intstr.IntOrString
				maxUnavailablePool1 intstr.IntOrString

				namePool2           string
				minPool2            int
				maxPool2            int
				maxSurgePool2       intstr.IntOrString
				maxUnavailablePool2 intstr.IntOrString

				vswitchZone1 string
				vswitchZone2 string
				zone1        string
				zone2        string

				workerPoolHash1 string
				workerPoolHash2 string

				shootVersionMajorMinor string
				shootVersion           string
				scheme                 *runtime.Scheme
				decoder                runtime.Decoder
				clusterWithoutImages   *extensionscontroller.Cluster
				cluster                *extensionscontroller.Cluster
				w                      *extensionsv1alpha1.Worker
			)

			BeforeEach(func() {
				namespace = "shoot--foobar--alicloud"
				cloudProfileName = "alicloud"

				region = "china"
				alicloudAccessKeyID = "access-key-id"
				alicloudAccessKeySecret = "secret-access-key"

				machineImageName = "my-os"
				machineImageVersion = "123"
				machineImageID = "ami-123456"

				instanceChargeType = "PostPaid"
				internetChargeType = "PayByTraffic"
				internetMaxBandwidthIn = 5
				internetMaxBandwidthOut = 5
				spotStrategy = "NoSpot"

				machineType = "large"
				userData = []byte("some-user-data")
				securityGroupID = "sg-12345"
				keyName = "my-ssh-key"

				volumeType = "normal"
				volumeSize = 20

				namePool1 = "pool-1"
				minPool1 = 5
				maxPool1 = 10
				maxSurgePool1 = intstr.FromInt(3)
				maxUnavailablePool1 = intstr.FromInt(2)

				namePool2 = "pool-2"
				minPool2 = 30
				maxPool2 = 45
				maxSurgePool2 = intstr.FromInt(10)
				maxUnavailablePool2 = intstr.FromInt(15)

				vswitchZone1 = "vswitch-acbd1234"
				vswitchZone2 = "vswitch-4321dbca"
				zone1 = region + "a"
				zone2 = region + "b"

				shootVersionMajorMinor = "1.2"
				shootVersion = shootVersionMajorMinor + ".3"

				clusterWithoutImages = &extensionscontroller.Cluster{
					Shoot: &gardencorev1beta1.Shoot{
						Spec: gardencorev1beta1.ShootSpec{
							Kubernetes: gardencorev1beta1.Kubernetes{
								Version: shootVersion,
							},
						},
					},
				}

				cloudProfileConfig := &apiv1alpha1.CloudProfileConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: apiv1alpha1.SchemeGroupVersion.String(),
						Kind:       "CloudProfileConfig",
					},
					MachineImages: []apiv1alpha1.MachineImages{
						{
							Name: machineImageName,
							Versions: []apiv1alpha1.MachineImageVersion{
								{
									Version: machineImageVersion,
									Regions: []apiv1alpha1.RegionIDMapping{
										{
											Name: region,
											ID:   machineImageID,
										},
									},
								},
							},
						},
					},
				}
				cloudProfileConfigJSON, _ := json.Marshal(cloudProfileConfig)
				cluster = &extensionscontroller.Cluster{
					CloudProfile: &gardencorev1beta1.CloudProfile{
						ObjectMeta: metav1.ObjectMeta{
							Name: cloudProfileName,
						},
						Spec: gardencorev1beta1.CloudProfileSpec{
							ProviderConfig: &gardencorev1beta1.ProviderConfig{
								RawExtension: runtime.RawExtension{
									Raw: cloudProfileConfigJSON,
								},
							},
						},
					},
					Shoot: clusterWithoutImages.Shoot,
				}

				w = &extensionsv1alpha1.Worker{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
					},
					Spec: extensionsv1alpha1.WorkerSpec{
						SecretRef: corev1.SecretReference{
							Name:      "secret",
							Namespace: namespace,
						},
						Region: region,
						InfrastructureProviderStatus: &runtime.RawExtension{
							Raw: encode(&api.InfrastructureStatus{
								VPC: api.VPCStatus{
									VSwitches: []api.VSwitch{
										{
											ID:      vswitchZone1,
											Purpose: "nodes",
											Zone:    zone1,
										},
										{
											ID:      vswitchZone2,
											Purpose: "nodes",
											Zone:    zone2,
										},
									},
									SecurityGroups: []api.SecurityGroup{
										{
											ID:      securityGroupID,
											Purpose: "nodes",
										},
									},
								},
								KeyPairName: keyName,
							}),
						},
						Pools: []extensionsv1alpha1.WorkerPool{
							{
								Name:           namePool1,
								Minimum:        minPool1,
								Maximum:        maxPool1,
								MaxSurge:       maxSurgePool1,
								MaxUnavailable: maxUnavailablePool1,
								MachineType:    machineType,
								MachineImage: extensionsv1alpha1.MachineImage{
									Name:    machineImageName,
									Version: machineImageVersion,
								},
								UserData: userData,
								Volume: &extensionsv1alpha1.Volume{
									Type: &volumeType,
									Size: fmt.Sprintf("%dGi", volumeSize),
								},
								Zones: []string{
									zone1,
									zone2,
								},
							},
							{
								Name:           namePool2,
								Minimum:        minPool2,
								Maximum:        maxPool2,
								MaxSurge:       maxSurgePool2,
								MaxUnavailable: maxUnavailablePool2,
								MachineType:    machineType,
								MachineImage: extensionsv1alpha1.MachineImage{
									Name:    machineImageName,
									Version: machineImageVersion,
								},
								UserData: userData,
								Volume: &extensionsv1alpha1.Volume{
									Type: &volumeType,
									Size: fmt.Sprintf("%dGi", volumeSize),
								},
								Zones: []string{
									zone1,
									zone2,
								},
							},
						},
					},
				}

				scheme = runtime.NewScheme()
				_ = api.AddToScheme(scheme)
				_ = apiv1alpha1.AddToScheme(scheme)
				decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()

				workerPoolHash1, _ = worker.WorkerPoolHash(w.Spec.Pools[0], cluster)
				workerPoolHash2, _ = worker.WorkerPoolHash(w.Spec.Pools[1], cluster)

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, clusterWithoutImages)
			})

			Describe("machine images", func() {
				var (
					defaultMachineClass map[string]interface{}
					machineDeployments  worker.MachineDeployments
					machineClasses      map[string]interface{}
				)

				BeforeEach(func() {
					defaultMachineClass = map[string]interface{}{
						"imageID":         machineImageID,
						"instanceType":    machineType,
						"region":          region,
						"securityGroupID": securityGroupID,
						"systemDisk": map[string]interface{}{
							"category": volumeType,
							"size":     volumeSize,
						},
						"instanceChargeType":      instanceChargeType,
						"internetChargeType":      internetChargeType,
						"internetMaxBandwidthIn":  internetMaxBandwidthIn,
						"internetMaxBandwidthOut": internetMaxBandwidthOut,
						"spotStrategy":            spotStrategy,
						"tags": map[string]string{
							fmt.Sprintf("kubernetes.io/cluster/%s", namespace):     "1",
							fmt.Sprintf("kubernetes.io/role/worker/%s", namespace): "1",
						},
						"secret": map[string]interface{}{
							"userData": string(userData),
						},
						"keyPairName": keyName,
					}

					var (
						machineClassPool1Zone1 = useDefaultMachineClass(defaultMachineClass, "vSwitchID", vswitchZone1, "zoneID", zone1)
						machineClassPool1Zone2 = useDefaultMachineClass(defaultMachineClass, "vSwitchID", vswitchZone2, "zoneID", zone2)
						machineClassPool2Zone1 = useDefaultMachineClass(defaultMachineClass, "vSwitchID", vswitchZone1, "zoneID", zone1)
						machineClassPool2Zone2 = useDefaultMachineClass(defaultMachineClass, "vSwitchID", vswitchZone2, "zoneID", zone2)

						machineClassNamePool1Zone1 = fmt.Sprintf("%s-%s-%s", namespace, namePool1, zone1)
						machineClassNamePool1Zone2 = fmt.Sprintf("%s-%s-%s", namespace, namePool1, zone2)
						machineClassNamePool2Zone1 = fmt.Sprintf("%s-%s-%s", namespace, namePool2, zone1)
						machineClassNamePool2Zone2 = fmt.Sprintf("%s-%s-%s", namespace, namePool2, zone2)

						machineClassWithHashPool1Zone1 = fmt.Sprintf("%s-%s", machineClassNamePool1Zone1, workerPoolHash1)
						machineClassWithHashPool1Zone2 = fmt.Sprintf("%s-%s", machineClassNamePool1Zone2, workerPoolHash1)
						machineClassWithHashPool2Zone1 = fmt.Sprintf("%s-%s", machineClassNamePool2Zone1, workerPoolHash2)
						machineClassWithHashPool2Zone2 = fmt.Sprintf("%s-%s", machineClassNamePool2Zone2, workerPoolHash2)
					)

					addNameAndSecretToMachineClass(machineClassPool1Zone1, alicloudAccessKeyID, alicloudAccessKeySecret, machineClassWithHashPool1Zone1)
					addNameAndSecretToMachineClass(machineClassPool1Zone2, alicloudAccessKeyID, alicloudAccessKeySecret, machineClassWithHashPool1Zone2)
					addNameAndSecretToMachineClass(machineClassPool2Zone1, alicloudAccessKeyID, alicloudAccessKeySecret, machineClassWithHashPool2Zone1)
					addNameAndSecretToMachineClass(machineClassPool2Zone2, alicloudAccessKeyID, alicloudAccessKeySecret, machineClassWithHashPool2Zone2)

					machineClasses = map[string]interface{}{"machineClasses": []map[string]interface{}{
						machineClassPool1Zone1,
						machineClassPool1Zone2,
						machineClassPool2Zone1,
						machineClassPool2Zone2,
					}}

					machineDeployments = worker.MachineDeployments{
						{
							Name:           machineClassNamePool1Zone1,
							ClassName:      machineClassWithHashPool1Zone1,
							SecretName:     machineClassWithHashPool1Zone1,
							Minimum:        worker.DistributeOverZones(0, minPool1, 2),
							Maximum:        worker.DistributeOverZones(0, maxPool1, 2),
							MaxSurge:       worker.DistributePositiveIntOrPercent(0, maxSurgePool1, 2, maxPool1),
							MaxUnavailable: worker.DistributePositiveIntOrPercent(0, maxUnavailablePool1, 2, minPool1),
						},
						{
							Name:           machineClassNamePool1Zone2,
							ClassName:      machineClassWithHashPool1Zone2,
							SecretName:     machineClassWithHashPool1Zone2,
							Minimum:        worker.DistributeOverZones(1, minPool1, 2),
							Maximum:        worker.DistributeOverZones(1, maxPool1, 2),
							MaxSurge:       worker.DistributePositiveIntOrPercent(1, maxSurgePool1, 2, maxPool1),
							MaxUnavailable: worker.DistributePositiveIntOrPercent(1, maxUnavailablePool1, 2, minPool1),
						},
						{
							Name:           machineClassNamePool2Zone1,
							ClassName:      machineClassWithHashPool2Zone1,
							SecretName:     machineClassWithHashPool2Zone1,
							Minimum:        worker.DistributeOverZones(0, minPool2, 2),
							Maximum:        worker.DistributeOverZones(0, maxPool2, 2),
							MaxSurge:       worker.DistributePositiveIntOrPercent(0, maxSurgePool2, 2, maxPool2),
							MaxUnavailable: worker.DistributePositiveIntOrPercent(0, maxUnavailablePool2, 2, minPool2),
						},
						{
							Name:           machineClassNamePool2Zone2,
							ClassName:      machineClassWithHashPool2Zone2,
							SecretName:     machineClassWithHashPool2Zone2,
							Minimum:        worker.DistributeOverZones(1, minPool2, 2),
							Maximum:        worker.DistributeOverZones(1, maxPool2, 2),
							MaxSurge:       worker.DistributePositiveIntOrPercent(1, maxSurgePool2, 2, maxPool2),
							MaxUnavailable: worker.DistributePositiveIntOrPercent(1, maxUnavailablePool2, 2, minPool2),
						},
					}

				})

				It("should return the expected machine deployments for profile image types", func() {
					workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

					expectGetSecretCallToWork(c, alicloudAccessKeyID, alicloudAccessKeySecret)

					// Test workerDelegate.DeployMachineClasses()

					chartApplier.
						EXPECT().
						ApplyChart(
							context.TODO(),
							filepath.Join(alicloud.InternalChartsPath, "machineclass"),
							namespace,
							"machineclass",
							machineClasses,
							nil,
						).
						Return(nil)

					err := workerDelegate.DeployMachineClasses(context.TODO())
					Expect(err).NotTo(HaveOccurred())

					// Test workerDelegate.GetMachineImages()
					machineImages, err := workerDelegate.GetMachineImages(context.TODO())
					Expect(machineImages).To(Equal(&apiv1alpha1.WorkerStatus{
						TypeMeta: metav1.TypeMeta{
							APIVersion: apiv1alpha1.SchemeGroupVersion.String(),
							Kind:       "WorkerStatus",
						},
						MachineImages: []apiv1alpha1.MachineImage{
							{
								Name:    machineImageName,
								Version: machineImageVersion,
								ID:      machineImageID,
							},
						},
					}))
					Expect(err).NotTo(HaveOccurred())

					// Test workerDelegate.GenerateMachineDeployments()

					result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(machineDeployments))
				})
			})

			It("should fail because the secret cannot be read", func() {
				c.EXPECT().
					Get(context.TODO(), gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
					Return(fmt.Errorf("error"))

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the version is invalid", func() {
				expectGetSecretCallToWork(c, alicloudAccessKeyID, alicloudAccessKeySecret)

				clusterWithoutImages.Shoot.Spec.Kubernetes.Version = "invalid"
				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the infrastructure status cannot be decoded", func() {
				expectGetSecretCallToWork(c, alicloudAccessKeyID, alicloudAccessKeySecret)

				w.Spec.InfrastructureProviderStatus = &runtime.RawExtension{}

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the security group cannot be found", func() {
				expectGetSecretCallToWork(c, alicloudAccessKeyID, alicloudAccessKeySecret)

				w.Spec.InfrastructureProviderStatus = &runtime.RawExtension{
					Raw: encode(&api.InfrastructureStatus{
						VPC: api.VPCStatus{},
					}),
				}

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the machine image cannot be found", func() {
				expectGetSecretCallToWork(c, alicloudAccessKeyID, alicloudAccessKeySecret)

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, clusterWithoutImages)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the vswitch id cannot be found", func() {
				expectGetSecretCallToWork(c, alicloudAccessKeyID, alicloudAccessKeySecret)

				w.Spec.InfrastructureProviderStatus = &runtime.RawExtension{
					Raw: encode(&api.InfrastructureStatus{
						VPC: api.VPCStatus{
							VSwitches: []api.VSwitch{},
							SecurityGroups: []api.SecurityGroup{
								{
									ID:      securityGroupID,
									Purpose: "nodes",
								},
							},
						},
					}),
				}

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail because the volume size cannot be decoded", func() {
				expectGetSecretCallToWork(c, alicloudAccessKeyID, alicloudAccessKeySecret)

				w.Spec.Pools[0].Volume.Size = "not-decodeable"

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})
})

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}

func expectGetSecretCallToWork(c *mockclient.MockClient, alicloudAccessKeyID, alicloudAccessKeySecret string) {
	c.EXPECT().
		Get(context.TODO(), gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
		DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
			secret.Data = map[string][]byte{
				alicloud.AccessKeyID:     []byte(alicloudAccessKeyID),
				alicloud.AccessKeySecret: []byte(alicloudAccessKeySecret),
			}
			return nil
		})
}

func useDefaultMachineClass(def map[string]interface{}, keyValues ...interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(def)+1)

	for k, v := range def {
		out[k] = v
	}

	for i := 0; i < len(keyValues); i += 2 {
		out[keyValues[i].(string)] = keyValues[i+1]
	}

	return out
}

func addNameAndSecretToMachineClass(class map[string]interface{}, alicloudAccessKeyID, alicloudAccessKeySecret, name string) {
	class["name"] = name
	class["secret"].(map[string]interface{})[alicloud.AccessKeyID] = alicloudAccessKeyID
	class["secret"].(map[string]interface{})[alicloud.AccessKeySecret] = alicloudAccessKeySecret
}
