/*
Copyright 2021 Absa Group Limited

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"regexp"

	k8gbv1beta1 "github.com/AbsaOSS/k8gb/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	types "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	externaldns "sigs.k8s.io/external-dns/endpoint"
)

func (r *GslbReconciler) updateGslbStatus(gslb *k8gbv1beta1.Gslb) error {
	var err error

	gslb.Status.ServiceHealth, err = r.getServiceHealthStatus(gslb)
	if err != nil {
		return err
	}

	err = r.Metrics.UpdateIngressHostsPerStatusMetric(gslb, gslb.Status.ServiceHealth)
	if err != nil {
		return err
	}

	gslb.Status.HealthyRecords, err = r.getHealthyRecords(gslb)
	if err != nil {
		return err
	}

	gslb.Status.GeoTag = r.Config.ClusterGeoTag

	err = r.Metrics.UpdateHealthyRecordsMetric(gslb, gslb.Status.HealthyRecords)
	if err != nil {
		return err
	}

	err = r.Status().Update(context.TODO(), gslb)
	return err
}

func (r *GslbReconciler) getServiceHealthStatus(gslb *k8gbv1beta1.Gslb) (map[string]string, error) {
	serviceHealth := make(map[string]string)
	for _, rule := range gslb.Spec.Ingress.Rules {
		for _, path := range rule.HTTP.Paths {
			service := &corev1.Service{}
			finder := client.ObjectKey{
				Namespace: gslb.Namespace,
				Name:      path.Backend.ServiceName,
			}
			err := r.Get(context.TODO(), finder, service)
			if err != nil {
				if errors.IsNotFound(err) {
					serviceHealth[rule.Host] = "NotFound"
					continue
				}
				return serviceHealth, err
			}

			endpoints := &corev1.Endpoints{}

			nn := types.NamespacedName{
				Name:      path.Backend.ServiceName,
				Namespace: gslb.Namespace,
			}

			err = r.Get(context.TODO(), nn, endpoints)
			if err != nil {
				return serviceHealth, err
			}

			serviceHealth[rule.Host] = "Unhealthy"
			if len(endpoints.Subsets) > 0 {
				for _, subset := range endpoints.Subsets {
					if len(subset.Addresses) > 0 {
						serviceHealth[rule.Host] = "Healthy"
					}
				}
			}
		}
	}
	return serviceHealth, nil
}

func (r *GslbReconciler) getHealthyRecords(gslb *k8gbv1beta1.Gslb) (map[string][]string, error) {

	dnsEndpoint := &externaldns.DNSEndpoint{}

	nn := types.NamespacedName{
		Name:      gslb.Name,
		Namespace: gslb.Namespace,
	}

	err := r.Get(context.TODO(), nn, dnsEndpoint)
	if err != nil {
		return nil, err
	}

	healthyRecords := make(map[string][]string)

	serviceRegex := regexp.MustCompile("^localtargets")
	for _, endpoint := range dnsEndpoint.Spec.Endpoints {
		local := serviceRegex.Match([]byte(endpoint.DNSName))
		if !local && endpoint.RecordType == "A" {
			if len(endpoint.Targets) > 0 {
				healthyRecords[endpoint.DNSName] = endpoint.Targets
			}
		}
	}

	return healthyRecords, nil
}
