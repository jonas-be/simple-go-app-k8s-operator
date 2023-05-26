/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SimpleGoAppFinalizer = "simplegoappfinalizer.simplegoapp.cluster.x-k8s.io"
)

// SimpleGoAppDeploymentSpec defines the desired state of SimpleGoAppDeployment
type SimpleGoAppDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SimpleGoAppDeployment. Edit simplegoappdeployment_types.go to remove/update
	ServerPort     int    `json:"serverPort,omitempty"`
	Replicas       int    `json:"replicas,omitempty"`
	ReturnValue    string `json:"returnValue,omitempty"`
	PodNameAsValue bool   `json:"podNameAsValue,omitempty"`
	NodePort       int32  `json:"nodePort,omitempty"`
}

// SimpleGoAppDeploymentStatus defines the observed state of SimpleGoAppDeployment
type SimpleGoAppDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	NodePort int32 `json:"nodePort,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SimpleGoAppDeployment is the Schema for the simplegoappdeployments API
type SimpleGoAppDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SimpleGoAppDeploymentSpec   `json:"spec,omitempty"`
	Status SimpleGoAppDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SimpleGoAppDeploymentList contains a list of SimpleGoAppDeployment
type SimpleGoAppDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SimpleGoAppDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SimpleGoAppDeployment{}, &SimpleGoAppDeploymentList{})
}
