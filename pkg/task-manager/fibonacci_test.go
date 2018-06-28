// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import "testing"

func TestFibonacciIndex(t *testing.T) {
	type args struct {
		index int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "0",
			args: args{1},
			want: 1,
		},
		{
			name: "1",
			args: args{1},
			want: 1,
		},
		{
			name: "2",
			args: args{2},
			want: 2,
		},
		{
			name: "6",
			args: args{6},
			want: 13,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FibonacciIndex(tt.args.index); got != tt.want {
				t.Errorf("FibonacciIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
