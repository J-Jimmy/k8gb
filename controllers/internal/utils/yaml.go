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

// Package utils implements common, reusable helpers
package utils

import (
	"encoding/json"

	k8gbv1beta1 "github.com/AbsaOSS/k8gb/api/v1beta1"
	yamlConv "github.com/ghodss/yaml"
)

// YamlToGslb takes yaml and returns Gslb object
func YamlToGslb(yaml []byte) (*k8gbv1beta1.Gslb, error) {
	// yamlBytes contains a []byte of my yaml job spec
	// convert the yaml to json
	jsonBytes, err := yamlConv.YAMLToJSON(yaml)
	if err != nil {
		return &k8gbv1beta1.Gslb{}, err
	}
	// unmarshal the json into the kube struct
	gslb := &k8gbv1beta1.Gslb{}
	err = json.Unmarshal(jsonBytes, &gslb)
	if err != nil {
		return &k8gbv1beta1.Gslb{}, err
	}
	return gslb, nil
}
