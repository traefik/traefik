// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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
	"github.com/snapcore/snapd/osutil"
)

const (
	encodedGenericAccount = `type: account
authority-id: canonical
account-id: generic
display-name: Generic
timestamp: 2017-07-27T00:00:00.0Z
username: generic
validation: certified
sign-key-sha3-384: -CvQKAwRQ5h3Ffn10FILJoEZUXOv6km9FwA80-Rcj-f-6jadQ89VRswHNiEB9Lxk

AcLDXAQAAQoABgUCWYuVIgAKCRDUpVvql9g3II66IACcoxSoX8+PQLa9TNuNBUs3bdTW6V5ZOdE8
vnziIg+yqu3qYfWHcRf1qu7K9Igv5lH3uM5jh2AHlndaoX4Qg1Rm9rOZCkRr1dDUmdRDBXN2pdTA
oydd0Ivpeai4ATbSZs11h50/vN/mxBwM6TzdGHqRNt6lvygAPe7VtfchSW/J0NsSIHr9SUeuIHkJ
C79DV27B+9/m8pnpKJo/Fv8nKGs4sMduKVjrj9Po3UhpZEQWf3I3SeDI5IE4TgoDe+O7neGUtT6W
D9wnMWLphC+rHbJguxXG/fmnUYiM2U8o4WVrs/fjF0zDRH7rY3tbLPbFXf2OD4qfOvS//VLQWeCK
KAgKhwz0d5CqaHyKSplywSvwO/dxlrqOjt39k3EjYxVuNS5UQk/BzPoDZD5maisCFm9JZwqBlWHP
6XTj8rhHSkNAPXezs2ZpVSsdtNYmpLLzWIFsAviuoMjYYDyL6jZrD4RBNrNOvSNQGLezB+eyI5DW
9vr2ppCw8zr49epPvJ4uqj/AILgr52zworl7v/27X67BOSoRMmE4AOnvjSJ8cN6Yt83AuEI4aZbP
DlF2Znqp8o/srtmJ3ZMpsjIsAqVhCeTU6eWXbYfNUlIMSmC6CDwQQzsukU4M6NEwUQbWddiM3iNL
FdeFsBscXg4Qm/0Y3PULriDoct+VpBUhzwVXG+Lj6rjtcX7n1C/7u9i/+WIBJ7jU4FBjwOdgpSCQ
DSCb0PgTM2PfbScFpn3KVYs0kT/Jc40Lpw6CUG9iUIdz5qlJzhbRiuhU8yjEg9q/5lWizAuxcP+P
anNhmNXsme46IJh7WnlzPAVMsToz8bWY01LC3t33pPGlRJo109PMbNK7reMIb4KFiL4Hy7gVmTj9
uydReVBUTZuMLRq1ShAJNScZ+HTpWruLoiC87Rf1++1KakahmtWYCdlJv/JSOyjSh8D9h0GEmqON
lKmzrNgQS8QhLh5uBcITN2Kt1UFGu2o9I8l0TgD5Uh9fG/R/A536fpcvIzOA/lhVttO6P9POwUVv
RIBZ3TpVOSzQ+ADpDexRUouPLPkgUwVBSctcgafMaj/w8DsULvlOYd3Sqq9a+zg6bZr9oPBKutUn
YkIUWmLW1OsdQ2eBS9dFqzJTAOELxNOUq37UGnIrMbk3Vn8hLK+S/+W9XL6WVxzmL1PT9FJZZ41p
KdaFV+mvrTfyoxuzXxkWbMwQkc56Ifn+IojbDwMI4FcTcl4dOeUrlnqwBJmTTwEhLVkYDvzYsVV9
4joFUWhp10JMm3lO+3596m0kYWMhyvGfYnH7QcQ3GtMAz82yRHc1X+seeWyD/aIjlHYNYfaJ5Ogs
VC76lXi7swMtA9jV5FJIGmQufLo9f93NSYxqwpa8
`

	encodedGenericModelsAccountKey = `type: account-key
authority-id: canonical
public-key-sha3-384: d-JcZF9nD9eBw7bwMnH61x-bklnQOhQud1Is6o_cn2wTj8EYDi9musrIT9z2MdAa
account-id: generic
name: models
since: 2017-07-27T00:00:00.0Z
body-length: 717
sign-key-sha3-384: -CvQKAwRQ5h3Ffn10FILJoEZUXOv6km9FwA80-Rcj-f-6jadQ89VRswHNiEB9Lxk

AcbBTQRWhcGAARAAoRakbLAMMoPeMz5MLCzAR6ALu/xxP9PuCdkknHH5lJrKE2adFj22DMwjWKj6
0pZU1Ushv4r7eb1NmFfl7a6Pz5ert+O5Qt53feK30+yiZF+Pgsx46SVTGy8QvicxhDhChdJ7ugW2
Vbz8dXDT9gv1E5hLl2BiuxxZHtMMTitO3bCtQcM/YwUeFljZZYd1FwxtgolnA5IUcHomIEQ5Xw6X
dCYGNkVjenb8aLBfi/ZZ84LHQjSbo3b87KP7syeEH2uuFJ2W8ZwGfUCll84gF+lYiLO6BQk8psIR
aRqnPfdjeuYg0ZLhdNV2Gu6GTNYMSrGLJ4vafAoIoMOifeIfK/DjN0XpfUIYwrM3UIvssEaLyE0L
i30PN5bpmmyfj5EDkJj9DqHzBly1kc20ciEtVCwOUijhQr4UjjfPiJFyed1/yndY1z/L85iATcsb
mwAw/wOyHKge/mlVztXV2H8DywcLV8Kbo5/ZZzcdKhDgL9URosQ5bMeYDPWwPsS02exHFl150dpR
p6MmeSCFPiQQjDrM3nWXLv/uemBE1IgX5q2eW6kJbSvRn519O3OrFEs2NBMEgvE3mIvewNlxFbDj
96Oj54Zh3rVtYu/g9yo2Bb2uf9gpOGS6TxrqN3aP5FigZzxkMCGFG8UOOFI7k2eQjMd8va5V8JTZ
ijWZgBjDB1YuQ1MAEQEAAQ==

AcLDXAQAAQoABgUCWYuUigAKCRDUpVvql9g3IOobH/wLm7sfLu3A/QWrdrMB1xRe6JOKuOQoNEt0
Vhg8q4MgOt1mxPzBUMGBJCcq9EiTYaUT4eDXSJL1OKFgh42oK5uY+GLsPWamxBY1Rg6QoESjJPcS
2niwTOjjTdpIrZ5M3pKRmxTxT+Wsq9j+1t4jvy/baI6+uO6KQh0UIMyOEhG+uJ8aJ2OcF3uV5gtF
fL1Y4Jr1Ir/4B2K7s8OhlrO1Yw3woB+YIkOjJ6oAOfQx5B/p1vK4uXOCIZarcfYX4XOhNgvPGaeL
O+NHk3GwTmEBngs49E8zq8ii8OoqIT6YzUd4taqHvZD4inTlw6MKGld7myCbZVZ3b0NXosplwYXa
jVL9ZBWTJukcIs4jEJ0XkTEuwvOpiGbtXdmDDlOSYkhZQdmQn3CIveGLRFa6pCi9a/jstyB+4sgk
MnwmJxEg8L3i1OvjgUM8uexCfg4cBVP9fCKuaC26uAXUiiHz7mIZhVSlLXHgUgMn5jekluPndgRZ
D2mGG0WscTMTb9uOpbLo6BWCwM7rGaZQgVSZsIj1cise05fjGpOozeqDhG25obcUXxhIUStztc9t
Z9MwSz9xdsUqV8XztEhkgfc7dh2fPWluiE9hLrdzyoU1xE6syujm8HE+bIJnDFYoE/Kw6WqIEm/3
mWhnOmi9uZsMBErKZKO4sqcLfR/zIn2Lx0ivg/yZzHHnDY5hwdrhQtn+AHCb+QJ9AJVte9hI+kt+
Fv8neohiMTCY8XxjrdB3QBPGesVsIMI5zAd14X4MqNKBYb4Ucg8YCIj7WLkQHbHO1GQwhPY8Tl9u
QqysZo/WnLVuvaruEBsBBGUJ7Ju5GtFKdWMdoH3YQmYHdxxxK37NPqBY70OrTSFJU5QT6PGFSvif
aMDg0X/aRj2uE3vgTI5hdqI4JYv1Mt1gYOPv4AMx/o/2q9dVENFYMTXcYBITMScUVV8NzmH8SNge
w7AWUPlQvWGZbTz62lYXHuUX1cdzz37B0LrEjh1ZC1V8emzfkLzEFYP/qUk1c4NjKsTjj5d463Gq
cn31Mr83tt5l7HWwP8bvTMIj98bOIJapsncGOzPYhs8cjZeOy0Q7EcvHjGRrj26CGWZacT3f0A0e
kb66ocAxV4nH1FDsfn8KdLKFgmSmW6SXkD2nqY94/pommJzUBF6s54DijZMXqHRwIRyPA8ymrCGt
t4shJh7dobC8Tg6RA84Bf9HkeqI97PQYFYMuNX0U59x2s0IQsOAYjH53NIf/jSPC4GDvLs7k+O76
R2PJK1VN6/ckJZAb3Rum5Ak5sbLTpRAVHIAVU1NAjHc5lYUHhxXJmJsbw6Jawb9Xb3T96s+WdD3Y
062upMY95pr0ZPf1tVGgzpcVCEw7yAOw+SkMksx+
`

	encodedGenericClassicModel = `type: model
authority-id: generic
series: 16
brand-id: generic
model: generic-classic
classic: true
timestamp: 2017-07-27T00:00:00.0Z
sign-key-sha3-384: d-JcZF9nD9eBw7bwMnH61x-bklnQOhQud1Is6o_cn2wTj8EYDi9musrIT9z2MdAa

AcLBXAQAAQoABgUCWYuXiAAKCRAdLQyY+/mCiST0D/0XGQauzV2bbTEy6DkrR1jlNbI6x8vfIdS8
KvEWYvzOWNhNlVSfwNOkFjs3uMHgCO6/fCg03wGXTyV9D7ZgrMeUzWrYp6EmXk8/LQSaBnff86XO
4/vYyfyvEYavhF0kQ6QGg8Cqr0EaMyw0x9/zWEO/Ll9fH/8nv9qcQq8N4AbebNvNxtGsCmJuXpSe
2rxl3Dw8XarYBmqgcBQhXxRNpa6/AgaTNBpPOTqgNA8ZtmbZwYLuaFjpZP410aJSs+evSKepy/ce
+zTA7RB3384YQVeZDdTudX2fGtuCnBZBAJ+NYlk0t8VFXxyOhyMSXeylSpNSx4pCqmUZRyaf5SDS
g1XxJet4IP0stZH1SfPOwc9oE81/bJlKsb9QIQKQRewvtUCLfe9a6Vy/CYd2elvcWOmeANVrJK0m
nRaz6VBm09RJTuwUT6vNugXSOCeF7W3WN1RHJuex0zw+nP3eCehxFSr33YrVniaA7zGfjXvS8tKx
AINNQB4g2fpfet4na6lPPMYM41WHIHPCMTz/fJQ6dZBSEg6UUZ/GiQhGEfWPBteK7yd9pQ8qB3fj
ER4UvKnR7hcVI26e3NGNkXP5kp0SFCkV5NQs8rzXzokpB7p/V5Pnqp3Km6wu45cU6UiTZFhR2IMT
l+6AMtrS4gDGHktOhwfmOMWqmhvR/INF+TjaWbsB6g==
`
)

var (
	genericAssertions        []asserts.Assertion
	genericStagingAssertions []asserts.Assertion
	genericExtraAssertions   []asserts.Assertion

	genericClassicModel         *asserts.Model
	genericStagingClassicModel  *asserts.Model
	genericClassicModelOverride *asserts.Model
)

func init() {
	genericAccount, err := asserts.Decode([]byte(encodedGenericAccount))
	if err != nil {
		panic(fmt.Sprintf(`cannot decode "generic"'s account: %v`, err))
	}
	genericModelsAccountKey, err := asserts.Decode([]byte(encodedGenericModelsAccountKey))
	if err != nil {
		panic(fmt.Sprintf(`cannot decode "generic"'s "models" account-key: %v`, err))
	}

	genericAssertions = []asserts.Assertion{genericAccount, genericModelsAccountKey}

	a, err := asserts.Decode([]byte(encodedGenericClassicModel))
	if err != nil {
		panic(fmt.Sprintf(`cannot decode "generic"'s "generic-classic" model: %v`, err))
	}
	genericClassicModel = a.(*asserts.Model)
}

// Generic returns a copy of the current set of predefined assertions for the 'generic' authority as used by Open.
func Generic() []asserts.Assertion {
	generic := []asserts.Assertion(nil)
	if !osutil.GetenvBool("SNAPPY_USE_STAGING_STORE") {
		generic = append(generic, genericAssertions...)
	} else {
		generic = append(generic, genericStagingAssertions...)
	}
	generic = append(generic, genericExtraAssertions...)
	return generic
}

// InjectGeneric injects further predefined assertions into the set used Open.
// Returns a restore function to reinstate the previous set. Useful
// for tests or called globally without worrying about restoring.
func InjectGeneric(extra []asserts.Assertion) (restore func()) {
	prev := genericExtraAssertions
	genericExtraAssertions = make([]asserts.Assertion, len(prev)+len(extra))
	copy(genericExtraAssertions, prev)
	copy(genericExtraAssertions[len(prev):], extra)
	return func() {
		genericExtraAssertions = prev
	}
}

// GenericClassicModel returns the model assertion for the "generic"'s "generic-classic" fallback model.
func GenericClassicModel() *asserts.Model {
	if genericClassicModelOverride != nil {
		return genericClassicModelOverride
	}
	if !osutil.GetenvBool("SNAPPY_USE_STAGING_STORE") {
		return genericClassicModel
	} else {
		return genericStagingClassicModel
	}
}

// MockGenericClassicModel mocks the predefined generic-classic model returned by GenericClassicModel.
func MockGenericClassicModel(mod *asserts.Model) (restore func()) {
	prevOverride := genericClassicModelOverride
	genericClassicModelOverride = mod
	return func() {
		genericClassicModelOverride = prevOverride
	}
}
