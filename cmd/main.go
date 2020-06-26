/*******************************************************************************
* Copyright 2020 VMware, Inc.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*******************************************************************************/

package main

import (
	"github.com/edgexfoundry/ev-charger/driver"

	"github.com/edgexfoundry/device-sdk-go/pkg/startup"
	device "github.com/edgexfoundry/ev-charger"
)

const (
	serviceName string = "ev-charger"
)

func main() {
	sd := driver.NewProtocolDriver()
	startup.Bootstrap(serviceName, device.Version, sd)
}
