/*
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
 */

package endpoints

import (
	"fmt"
	"strings"
)

const keyFormatter = "%s::%s"

var endpointMapping = make(map[string]string)

func AddEndpointMapping(regionId, productId, endpoint string) (err error) {
	key := fmt.Sprintf(keyFormatter, strings.ToLower(regionId), strings.ToLower(productId))
	endpointMapping[key] = endpoint
	return nil
}

type MappingResolver struct {
}

func (resolver *MappingResolver) TryResolve(param *ResolveParam) (endpoint string, support bool, err error) {
	key := fmt.Sprintf(keyFormatter, strings.ToLower(param.RegionId), strings.ToLower(param.Product))
	endpoint, contains := endpointMapping[key]
	return endpoint, contains, nil
}
