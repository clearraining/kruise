/*
Copyright 2020 The Kruise Authors.

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

package expectations

import (
	"strconv"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ResourceVersionExpectation interface {
	Expect(obj metav1.Object)
	Observe(obj metav1.Object)
	IsSatisfied(obj metav1.Object) bool
	Delete(obj metav1.Object)
}

func NewResourceVersionExpectation() ResourceVersionExpectation {
	return &realResourceVersionExpectation{objectVersions: make(map[types.UID]string, 100)}
}

type realResourceVersionExpectation struct {
	sync.Mutex
	objectVersions map[types.UID]string
}

func (r *realResourceVersionExpectation) Expect(obj metav1.Object) {
	r.Lock()
	defer r.Unlock()
	if isResourceVersionNewer(r.objectVersions[obj.GetUID()], obj.GetResourceVersion()) {
		r.objectVersions[obj.GetUID()] = obj.GetResourceVersion()
	}
}

func (r *realResourceVersionExpectation) Observe(obj metav1.Object) {
	r.Lock()
	defer r.Unlock()
	if isResourceVersionNewer(r.objectVersions[obj.GetUID()], obj.GetResourceVersion()) {
		delete(r.objectVersions, obj.GetUID())
	}
}

func (r *realResourceVersionExpectation) IsSatisfied(obj metav1.Object) bool {
	r.Lock()
	defer r.Unlock()
	if isResourceVersionNewer(r.objectVersions[obj.GetUID()], obj.GetResourceVersion()) {
		delete(r.objectVersions, obj.GetUID())
	}
	_, existing := r.objectVersions[obj.GetUID()]
	return !existing
}

func (r *realResourceVersionExpectation) Delete(obj metav1.Object) {
	r.Lock()
	defer r.Unlock()
	delete(r.objectVersions, obj.GetUID())
}

func isResourceVersionNewer(old, new string) bool {
	if len(old) == 0 {
		return true
	}

	oldCount, err := strconv.ParseUint(old, 10, 64)
	if err != nil {
		return true
	}

	newCount, err := strconv.ParseUint(new, 10, 64)
	if err != nil {
		return false
	}

	return newCount >= oldCount
}
