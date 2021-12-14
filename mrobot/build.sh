#!/bin/bash

#
# Copyright © 2021 zibuyu28
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
set -e

echo "start to build mrobot"
go build -o mrobot ./cmd/mrobot.go

echo "start to build mrobot-linux"
i=$(uname)
if [[ $i = "Darwin" ]]; then
  GOOS=linux GOARCH=amd64 go build -o mrobot-linux cmd/mrobot.go
else
  go build -o mrobot-linux cmd/mrobot.go
fi
echo "start to build image for k8s:v1.0.0"
docker build -t k8s:1.0.0 -f Dockerfile.k8s .

echo "success!"
