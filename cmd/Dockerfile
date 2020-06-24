###########################################################################
# Copyright 2020 VMware, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
###########################################################################

ARG BASE=golang:1.13-alpine
FROM ${BASE} AS builder

WORKDIR /ev-charger

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright 2020 VMware, Inc'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

# add git for go modules
RUN apk add --update --no-cache make git

COPY . .

RUN make build

# Next image - Copy built Go binary into new workspace
FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2020: Intel'

ENV APP_PORT=45678
#expose command data port
EXPOSE $APP_PORT

WORKDIR /
COPY --from=builder /ev-charger/cmd/ev-charger /usr/local/bin/ev-charger
COPY --from=builder /ev-charger/cmd/res/configuration.toml /res/configuration.toml
COPY --from=builder /ev-charger/cmd/res/EV-ChargePoint.yaml /res/EV-ChargePoint.yaml
CMD [ "/usr/local/bin/ev-charger","-cp=consul.http://172.21.0.2:8500", "--registry", "--confdir=/res"]
