#!/bin/bash

# Copyright 2015 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"

kube::golang::setup_env

APIROOTS=${APIROOTS:-pkg/api pkg/apis pkg/watch staging/src/k8s.io/apimachinery/pkg/api staging/src/k8s.io/apimachinery/pkg/apis staging/src/k8s.io/apiserver/pkg staging/src/k8s.io/api staging/src/k8s.io/metrics/pkg/apis}
_tmp="${KUBE_ROOT}/_tmp"

cleanup() {
  rm -rf "${_tmp}"
}

trap "cleanup" EXIT SIGINT

cleanup
for APIROOT in ${APIROOTS}; do
  mkdir -p "${_tmp}/${APIROOT}"
  cp -a "${KUBE_ROOT}/${APIROOT}"/* "${_tmp}/${APIROOT}/"
done

KUBE_VERBOSE=3 "${KUBE_ROOT}/hack/update-generated-protobuf.sh"
for APIROOT in ${APIROOTS}; do
  TMP_APIROOT="${_tmp}/${APIROOT}"
  echo "diffing ${APIROOT} against freshly generated protobuf"
  ret=0
  diff -Naupr -I 'Auto generated by' -x 'zz_generated.*' "${KUBE_ROOT}/${APIROOT}" "${TMP_APIROOT}" || ret=$?
  cp -a "${TMP_APIROOT}"/* "${KUBE_ROOT}/${APIROOT}/"
  if [[ $ret -eq 0 ]]; then
    echo "${APIROOT} up to date."
  else
    echo "${APIROOT} is out of date. Please run hack/update-generated-protobuf.sh"
    exit 1
  fi
done
