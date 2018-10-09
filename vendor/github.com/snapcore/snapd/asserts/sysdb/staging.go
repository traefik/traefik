// -*- Mode: Go; indent-tabs-mode: t -*-
// +build withtestkeys withstagingkeys

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package sysdb

import (
	"fmt"

	"github.com/snapcore/snapd/asserts"
)

const (
	encodedStagingTrustedAccount = `type: account
authority-id: canonical
account-id: canonical
display-name: Canonical
timestamp: 2016-04-01T00:00:00.0Z
username: canonical
validation: certified
sign-key-sha3-384: e2r8on4LgdxQSaW5T8mBD5oC2fSTktTVxOYa5w3kIP4_nNF6L7mt6fOJShdOGkKu

AcLBXAQAAQoABgUCV640ggAKCRAHKljtl9kuLrQtEADBji8VwAuislurkFORTmcXV/DOkvyvAYEN
mB/MLniK4MlLX+RDncDBmF38IK9SRkxbwwJuKgvsjwsYJ3w1P7SGvVfNyU2hLRFtdxDMVC7+A9g3
N1VW9W+IOWmYeBgXiveqAlSJ9GUvLQiBgUWRBkbyAT6aLkSZrTSjxGRGW/uoNfjj+CbAR4HGbRnn
IOxDuQyw6rOXQZKfZvkD1NiH+0QzXLv0RivE8+V5uVN+ooUFRoVQmqbj7orvPS9iTY5AMVjCgfo0
UiWiN6NyCfDBDz0bZhIZlBU4JF5W0I/sEwsuYCxIhFi5uPNmQXqqb5d9Y3bsxIUdMR0+pai1A3eI
HQmYX12wCnb276R5Adz4iol19oKAR2Qf3VJBvPccdIFU7Qu5FOOihQdMRxULBBXGn1HQF1uW+ue3
ZQ3x6e8s3XjdDQE/kHCDUkmzhbk1SErgndg6Q1ipKJ+4G6dOc16s66bSFA4QzW53Y40NP0HRWxe2
tK9VOJ+z9GvGYp5H1ZXbbs78t9bUwL7L6z/eXM6BRho6YY9X7nImpByIkdcV47dCyVFol6NrM5NS
NSpdtRStGqo7tjPaBf86p2vLOAbwFUuaE3rwf5g/agz4S/v5G5E2tKmfQs6vGYrfVtlOzr8gEoXH
+/hOEC3wYEJjpXmFRjUjJwr0Fbej2TpoITpfzbySpg==
`
	encodedStagingRootAccountKey = `type: account-key
authority-id: canonical
revision: 3
public-key-sha3-384: e2r8on4LgdxQSaW5T8mBD5oC2fSTktTVxOYa5w3kIP4_nNF6L7mt6fOJShdOGkKu
account-id: canonical
name: staging-root
since: 2016-04-01T00:00:00.0Z
body-length: 717
sign-key-sha3-384: e2r8on4LgdxQSaW5T8mBD5oC2fSTktTVxOYa5w3kIP4_nNF6L7mt6fOJShdOGkKu

AcbBTQRWhcGAARAA4wh+b9nyRdZj9gNKuHz8BTNZsLOVv2VJseHBoMNc4aA8EgmLwMF/aP+q1tAQ
VOeynhfSecIK/2aWKKX+dmU/rfAbnbdHX1NT8OnG2z3qdYdqw1EreN8LcY4DBDfa1RNKcjFvBu+Q
jxpU289m1yUjjc7yHie84BoYRgDl0icar8KF7vKx44wNhzbca+lw4xGSA5gpDZ1i1smdxdpOSsUY
WT70ZcJBN1oyKiiiCJUNLwCPzaPsH1i3WwDSaGsbjl8gjf2+LFNFPwdsWRbn3RLlFcFbET2bFe5y
v6UN+0cSh9qLJeLR2h0WDaVBp5Gx4PAYAfpIIF8EH3YbvI8uuTmBza8Ni0yozOZ2cXCSdezLGW2m
b6itOq/taBhgl8gzhKqki9jAOWmDBeBIbe2rUuNJrfHVH8+lWTzuzJIcHSHeAjFG1xid+HOOsw0e
Ag3JMjJaqCGCp0Oc9/WBtHV6jB30jLzht5QjJZ6izIKswRrvt0nCowp74FZ1l1ekXZPhhkA5MBMb
AoTiz9UvRZAWBPa5gX4R7eaekGjCPWI8NpJ7pT3Xh3NIHIsjyf0JcysoH2V1+A9qT1LOCyczf1Uc
9d8PXap1zhhQuczZcnw7vAwLEIwldfp08x6klsgiP6jqIB4XKJCjBDu/gn682ydWzfLT8echVHpg
uI62X67Ns1ZbFWMAEQEAAQ==

AcLBXAQAAQoABgUCV86jSgAKCRAHKljtl9kuLpV6EADO8Q1WKJwoTfeIpBpQfDhdhqJLmW86Qrjq
P9ZsndN8eA4uSbo08yg9jxi6Q3J/A5QK6rhTz5Nu41frKVpgFr80BpIx8cHsY2dZNyKCm70Jjy4h
cxteK7mwdAzdWG/Dg7Nr4fhOmpepsh1gIXvjWhTkT226DIO6l45o6N2hMKKkWmqJYqVD6i7UE4Ed
xmC+IoluhnKGGwM6JpyOw0RViXbLjVDR58n4q1xmK7cFduMoLKszVY4/KGmKT8gA6D4pUOa62F84
Ejh6akRum7uqygBibYT/DP+KA+MhHvpQ8XZem7IVIEnMJr7U2gde3brbVr0oiCl7FzfiBNy6qw92
cTsE8o3JV0Lc106SWU28GuWPgyXjoH8imzSmWlpQtlPlKEDwMQt31XDKUKp0ZKiEax3cQ6VjMv1f
PV3bHNjD+tBq5e1xm/UWyGu7J2N4VPLgUK7F4TPUJk5lwKjmII8KD3KA/IeHnZVN6vmC2nKfhGvw
+rJllQQ0IWY9RfIdzFHpVvthe48g27ok5yEgovAc/s7xWZ6CBgyzYWLQMNFvENj04CzGvxirKwuJ
Fy5UJIEKB0j0R2qnCz6HZkyQrUsz5HiIIlks18FfOZwuIc4GGPbwwQBoXW7a6KQg0aa62BPj5Iww
3w60rtTSUsjINkZ/GXLodfzPglOl6VLF7bWx2hGesg==
`
	encodedStagingGenericAccount = `type: account
authority-id: canonical
account-id: generic
display-name: Generic
timestamp: 2017-07-27T00:00:00.0Z
username: generic
validation: certified
sign-key-sha3-384: e2r8on4LgdxQSaW5T8mBD5oC2fSTktTVxOYa5w3kIP4_nNF6L7mt6fOJShdOGkKu

AcLBXAQAAQoABgUCWXmmFAAKCRAHKljtl9kuLkAWD/98LgECwAN8S09o4aEFpdGXgWpx8z58wl6T
5mZVDyYpCV9ugC2DqBqGQxp4X1P7Wn9+weXw8nmL7IywVn/hCVHJOmBLJSr3wLjpVBY9RrIHYoXi
k9W7IFo4ggw1j1FRLg2tKk81MnK0fK/Qws9OXzilDir5R2bQ/E0sodGW3NpbwtbpkY/BtP6YPoJ/
1+205KG5m6oG8y6mf74bjMGfJ+iFFpIDayIpXl+YTkJ25BOVGcuC66cIrmdc63rBIHL2tU/3GUMB
xZGiyG9Fuli1uV4ALhN9j43hxAtVwXOn/qgOiN8TGQz3OvlVUXTuFVmkdvCdfT2XHrJjFmEs9SlL
u2EEmvaNFJ61lQG/VrN6O0BswenTlIO0tTFe126o/cTmKg8/ga4v2WjMlcOCzfu+cIZIzTTnn4Le
iXdQ6+c3QN+Co4SI0UvgJ4nGWQ9W+4q4xVJTliKTzK2BZ40vHUi51rMC/puqsMpnAbHSn4iy8vpf
CyJh7jyuITPEzfpurNMb+VD+1Brd2DJCVnlwQq+rzNerXd5xcHCdZsfX+ATukHgYTZWa467ZEFhI
Bk1xUWAYs8r2JDFb5YPtZuW7Vt1UUpFdx6DroL6OODvZ6mDUtsOa8nm7G1l4uRJtqunplPyCDjnL
aQhlAouLMltWeGITO+5jePHJKTnYQAFEvo0WIgEYpA==
`
	encodedStagingGenericModelsAccountKey = `type: account-key
authority-id: canonical
public-key-sha3-384: 93jDrIGOXymDg9BPCLES5mAr6aGXU7e0wwXeJlIYIWbUzM_kB81CiqX7cTlB9Y1z
account-id: generic
name: models
since: 2017-07-27T00:00:00.0Z
body-length: 717
sign-key-sha3-384: e2r8on4LgdxQSaW5T8mBD5oC2fSTktTVxOYa5w3kIP4_nNF6L7mt6fOJShdOGkKu

AcbBTQRWhcGAARAAxcyFC13COEmIwWwLsjp4AAILhWSp8/dQ6cOzY3T7tqqoSn9iKyidpJTfrtml
DKHZe0zC10fog2Mvp1AO7dNqK9kHUdCQE+YatHmkm1a3QoZqwJsj77w09Q+l1uvDjrfrF0S/KcYa
0hfDZQ+51T1msbatWN2qU42dX280IMV+zo1GpKK8z6br4glY2tki4CJokVAAt+bl4bBqDZ4EoBYe
9CsACmNhw2d/fOAlis2jwG3tWXMORX9FcGRx/COasvRb7rjA0DJfKxOnTw7uC0UjDUB6bU6O0smS
Q5oK3V5fJIAcMBXNe5MkdTKGLY61hTFLiw4F6MkrM3O8dnXtexCojV+QtROTIM4R2dJTOv7r2s/9
KT4wIQmOcjMEWxyq1H1rqjCHBjGnKa6GC1j/4NwqlxUEiqZMYs12px9ypeEqjL3tKURCanPOlwXO
p2E1i+V53XznnS3RA7I6Aa37w/9clTJk5vzVT8G6+k6xsB9zwKYOsipG0zHjyuW9Qtkd15bA0Iv9
MrZGE4U7RwEnt4jBa98rcLs1sCkJEau0hEU4MiPyqi8XL2b/TtPnCwN8rQRVQvakzGw83Ol/B8ZI
2OGu0aB6HAWbdy81yXIUES9ZtH7nK5X7dSdJu92wXBMOyel9cryHzlYFjSlPKyqRx2lsYk4K6Hiq
VRY3L12yjXkcHcsAEQEAAQ==

AcLBXAQAAQoABgUCWXmevAAKCRAHKljtl9kuLltVEADAfpBlY4b4oImKPq8Pp6UKFjgcMVjJLcSI
EOfAMygIaZwzNSuOh2wPRBAMMZlcFlBEFLfGbh7R2RG1/R7PSR4q+gMZZ3qJ5QUjUUuGnkSfCLhK
jVtPlX8qdPWTdEgUeTEKNzHogP0MiIChHdeuv/iQ9fSgdw/lBZsblKAdrHv00ZQHup9XGWdZ4Fnb
cSiK9tZ3/nZAG18PErEh7wJntwygqcjScS0jTSa5BecQoy8O6wxKafQgxuixdHw+dt6sa34qzwel
ROb1VmNcmMGsv2YuPsRcqjgvL7drDzXRRcYhmiSCUFhGPx3RY0UWO9G9Pzok64l/1D7o6Xah3h4D
oxkepM5JTAiy165kfQzEFGMtvlv0d3mOCLMWqJzjzhhn/bPcoh5MO/PhpSR1y71tjjWtKR4SD1K/
feR+KE73gEgqmHLss3TF/O2RxvbV15W0paxmiffUyeE/uQ1p5ddmBwke/gM/OOgUA4G3g9vgTeaQ
YVFCD0h75mX9GylVXUMmxlYSRVO59JsxCpWSqXWh4xkigbBZSAOPn6vkM/4nxZLf/ufVyNqnwo2r
Si/lWgk/lJoHaTVsqm9n0DxiR8lH54eFghprkN+KgGFMKlY127n50CEwqG1j4gKfjxmRycgKbx5O
a0IGmGjVnVmGOdX9wpg6fBHbhczXZtId02Q7yzF87A==
`

	encodedStagingGenericClassicModel = `type: model
authority-id: generic
series: 16
brand-id: generic
model: generic-classic
classic: true
timestamp: 2017-07-27T00:00:00.0Z
sign-key-sha3-384: 93jDrIGOXymDg9BPCLES5mAr6aGXU7e0wwXeJlIYIWbUzM_kB81CiqX7cTlB9Y1z

AcLBXAQAAQoABgUCWXnYbAAKCRDqhmvwxUsbelvOD/4qxiDs4blJoRSXmzvsKTyM/Z2QuLw9bqUj
QXKoCSB78ATFwr01kvqJMzwJ1eT4zKOajUERKPN9fN1af0w07DoYG5bt/Pb7s/UFDmwIQg244wLI
lQ/NPCAm4SEvN1GEe0OxdCpMuPe+x++FvFtnF7CXJPmLdHln6A1eMhwyxGX+el1QxhiR+mWLCCNp
B4ndjh154H5SXRw1lmUiYdE/kCsOqGeZ5ljTni+Rh8xDYxmVCthrLCUVtHhMVrKeylDwwS7Sf/HV
GY9r/C9r07xRom06bBN/vQwdoLzGuU3SS7UsN0Ud95tJhAUtP5jW1dN8otviMcOAdtj7jTwSX4FY
pdgmkldjaCRaHxBA923cjGgl98LCjbdG5KmmKoT6DTb3AyFOT2XwlRl/MaRJBK2Tp1nVNZDjLY4j
VfRETt17ZCONt3yn/OhQk8bV6EsdJvT2/nMlNejXgnMtLfbH8v6xWLKrLOVOjILVF5zgK8+z4+d2
ILIZupGooMouhddmcHem76lSnS+y75NMQXg5lBrUU2xAQRloWTw0oF+Hr5vcZkX5f4R/yH8Zz1Dt
+zRs2zqOK5hjdejhU5x/N3KSBLy+TUMk7JsdVv0nhdpJUKrFyGWn+YzBNE2GgEfPfXnkaU91/AD2
SWyt8kWVPmT3DCzs7u5IXYIVxcq4FjkmeU9sTrn88g==
`
)

func init() {
	stagingTrustedAccount, err := asserts.Decode([]byte(encodedStagingTrustedAccount))
	if err != nil {
		panic(fmt.Sprintf("cannot decode trusted assertion: %v", err))
	}
	stagingRootAccountKey, err := asserts.Decode([]byte(encodedStagingRootAccountKey))
	if err != nil {
		panic(fmt.Sprintf("cannot decode trusted assertion: %v", err))
	}
	trustedStagingAssertions = []asserts.Assertion{stagingTrustedAccount, stagingRootAccountKey}

	genericAccount, err := asserts.Decode([]byte(encodedStagingGenericAccount))
	if err != nil {
		panic(fmt.Sprintf(`cannot decode "generic"'s account: %v`, err))
	}
	genericModelsAccountKey, err := asserts.Decode([]byte(encodedStagingGenericModelsAccountKey))
	if err != nil {
		panic(fmt.Sprintf(`cannot decode "generic"'s "models" account-key: %v`, err))
	}

	genericStagingAssertions = []asserts.Assertion{genericAccount, genericModelsAccountKey}

	a, err := asserts.Decode([]byte(encodedStagingGenericClassicModel))
	if err != nil {
		panic(fmt.Sprintf(`cannot decode "generic"'s "generic-classic" model: %v`, err))
	}
	genericStagingClassicModel = a.(*asserts.Model)
}
