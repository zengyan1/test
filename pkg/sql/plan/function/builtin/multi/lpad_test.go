// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package multi

import (
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestLpadVarchar(t *testing.T) {
	cases := []struct {
		name      string
		vecs      []*vector.Vector
		wantBytes []byte
	}{
		{
			name:      "TEST01",
			vecs:      makeLpadVectors("hello", 1, "#", []int{1, 1, 1}),
			wantBytes: []byte("h"),
		},
		{
			name:      "TEST02",
			vecs:      makeLpadVectors("hello", 10, "#", []int{1, 1, 1}),
			wantBytes: []byte("#####hello"),
		},
		{
			name:      "TEST03",
			vecs:      makeLpadVectors("hello", 15, "#@&", []int{1, 1, 1}),
			wantBytes: []byte("#@&#@&#@&#hello"),
		},
		{
			name:      "TEST04",
			vecs:      makeLpadVectors("12345678", 10, "abcdefgh", []int{1, 1, 1}),
			wantBytes: []byte("ab12345678"),
		},
		{
			name:      "TEST05",
			vecs:      makeLpadVectors("hello", 0, "#@&", []int{1, 1, 1}),
			wantBytes: []byte(""),
		},
		{
			name:      "TEST06",
			vecs:      makeLpadVectors("hello", -1, "#@&", []int{1, 1, 1}),
			wantBytes: []byte(nil),
		},
		{
			name:      "Tx",
			vecs:      makeLpadVectors("hello", 1, "", []int{1, 1, 1}),
			wantBytes: []byte(""),
		},
		{
			name:      "Tx2",
			vecs:      makeLpadVectors("", 5, "x", []int{1, 1, 1}),
			wantBytes: []byte("xxxxx"),
		},
		{
			name:      "Tx3",
			vecs:      makeLpadVectors("??????", 10, "??????", []int{1, 1, 1}),
			wantBytes: []byte("??????????????????????????????"),
		},
		{
			name:      "tx4",
			vecs:      makeLpadVectors("hello", -1, "#@&", []int{0, 0, 0}),
			wantBytes: []byte(nil),
		},
		{
			name:      "tx5",
			vecs:      makeLpadVectors("hello", -1, "#@&", []int{0, 0, 1}),
			wantBytes: []byte(nil),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("hello", -1, "#@&", []int{0, 1, 0}),
			wantBytes: []byte(nil),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("hello", -1, "#@&", []int{1, 0, 0}),
			wantBytes: []byte(nil),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("hello", -1, "#@&", []int{1, 1, 0}),
			wantBytes: []byte(nil),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("hello", -1, "#@&", []int{1, 0, 1}),
			wantBytes: []byte(nil),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("hello", -1, "#@&", []int{0, 1, 1}),
			wantBytes: []byte(nil),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("hello", -1, "#@&", []int{1, 1, 1}),
			wantBytes: []byte(nil),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("a???", 15, "???", []int{1, 1, 1}),
			wantBytes: []byte("???????????????????????????????????????a???"),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("a???a", 15, "???a", []int{1, 1, 1}),
			wantBytes: []byte("???a???a???a???a???a???aa???a"),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("a???aa", 15, "???aa", []int{1, 1, 1}),
			wantBytes: []byte("???aa???aa???aa???aa???aa"),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("a???aaa", 15, "???aaa", []int{1, 1, 1}),
			wantBytes: []byte("???aaa???aaa???aa???aaa"),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("a???aaaa", 15, "???aaaa", []int{1, 1, 1}),
			wantBytes: []byte("???aaaa???aaaa???aaaa"),
		},
		{
			name:      "tx6",
			vecs:      makeLpadVectors("aaaaaaaa", 4, "bbb", []int{1, 1, 1}),
			wantBytes: []byte("aaaa"),
		},
	}

	proc := testutil.NewProcess()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lpad, err := Lpad(c.vecs, proc)
			if err != nil {
				t.Fatal(err)
			}
			if c.wantBytes == nil {
				ret := nulls.Contains(lpad.Nsp, 0)
				require.Equal(t, ret, true)
			} else {
				require.Equal(t, c.wantBytes, lpad.GetBytes(0))
			}

		})
	}

}

func makeLpadVectors(src string, length int64, pad string, nils []int) []*vector.Vector {
	vec := make([]*vector.Vector, 3)
	vec[0] = vector.NewConstString(types.T_varchar.ToType(), 1, src, testutil.TestUtilMp)
	vec[1] = vector.NewConstFixed(types.T_int64.ToType(), 1, length, testutil.TestUtilMp)
	vec[2] = vector.NewConstString(types.T_varchar.ToType(), 1, pad, testutil.TestUtilMp)
	for i, n := range nils {
		if n == 0 {
			nulls.Add(vec[i].Nsp, uint64(i))
		}
	}
	return vec
}
