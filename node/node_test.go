package node

import (
	"reflect"
	"testing"
)

func TestPreNode_MarshalJSON(t *testing.T) {
	type fields struct {
		ID  string
		Cnt int
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "simple_0",
			fields: fields{
				ID:  "1",
				Cnt: 0,
			},
			want: []byte(`["1",0]`),
		},
		{
			name: "simple_1",
			fields: fields{
				ID:  "100",
				Cnt: 200,
			},
			want: []byte(`["100",200]`),
		},
		{
			name: "escape",
			fields: fields{
				ID:  `"1000"`,
				Cnt: 1001,
			},
			want: []byte(`["\"1000\"",1001]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := PreNode{
				ID:   tt.fields.ID,
				Argc: tt.fields.Cnt,
			}
			got, err := n.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})

	}
}

func TestPreNode_UnmarshalJSON(t *testing.T) {
	type fields struct {
		ID  string
		Cnt int
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "simple_0",
			fields: fields{
				ID:  "1",
				Cnt: 0,
			},
			args: args{
				p: []byte(`["1",0]`),
			},
		},
		{
			name: "simple_1",
			fields: fields{
				ID:  "100",
				Cnt: 200,
			},
			args: args{
				p: []byte(`["100",200]`),
			},
		},
		{
			name: "escape",
			fields: fields{
				ID:  "1000",
				Cnt: 1001,
			},
			args: args{
				p: []byte(`["\"1000\"",1001]`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &PreNode{
				ID:   tt.fields.ID,
				Argc: tt.fields.Cnt,
			}
			if err := n.UnmarshalJSON(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
