# Copyright 2022 TriggerMesh Inc.
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

FROM golang:1.19 AS builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/project

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -a -o triggermesh-core ./cmd/core-controller

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /go/src/project/triggermesh-core /
COPY LICENSE EULA.pdf /licenses/

USER 65532:65532

ENTRYPOINT ["/triggermesh-core"]
