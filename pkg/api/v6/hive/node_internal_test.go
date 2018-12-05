package hive

import (
	"reflect"
	"testing"
)

func Test_NodeType(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    string
		wantErr bool
	}{
		{"Valid", "weeee", "weeee", false},
		{"Invalid", 90000, "", true},
		{"Missing", nil, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &node{Attributes: make(nodeAttributes)}

			if tt.value != nil {
				n.Attributes["nodeType"] = &nodeAttribute{
					ReportedValue: tt.value,
				}
			}

			got, err := n.NodeType()
			if (err != nil) != tt.wantErr {
				t.Errorf("n.NodeType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("n.NodeType() = %v, want %v", got, tt.want)
			}
		})
	}

}

func Test_node_attr(t *testing.T) {
	type fields struct {
		key   string
		value interface{}
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *nodeAttribute
	}{
		{"Present", fields{"example", "val"}, args{"example"}, &nodeAttribute{ReportedValue: "val"}},
		{"Missing", fields{"example", "val"}, args{"another"}, &nodeAttribute{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &node{
				Attributes: nodeAttributes{
					tt.fields.key: &nodeAttribute{
						ReportedValue: tt.fields.value,
					},
				},
			}

			if got := n.attr(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("node.attr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nodeAttribute(t *testing.T) {
	tests := []struct {
		name   string
		attr   *nodeAttribute
		want   interface{}
		wantOk bool
	}{
		{"ReportedValueString", &nodeAttribute{ReportedValue: nil}, nil, false},
		{"ReportedValueString", &nodeAttribute{ReportedValue: "tests"}, "tests", true},
		{"ReportedValueBool", &nodeAttribute{ReportedValue: nil}, nil, false},
		{"ReportedValueBool", &nodeAttribute{ReportedValue: true}, true, true},
		{"ReportedValueFloat", &nodeAttribute{ReportedValue: nil}, nil, false},
		{"ReportedValueFloat", &nodeAttribute{ReportedValue: 1.0}, 1.0, true},
		{"ReportedValueInt", &nodeAttribute{ReportedValue: nil}, nil, false},
		{"ReportedValueInt", &nodeAttribute{ReportedValue: -100}, int64(-100), true},
		{"ReportedValueUint", &nodeAttribute{ReportedValue: nil}, nil, false},
		{"ReportedValueUint", &nodeAttribute{ReportedValue: 100}, uint64(100), true},
		{"DisplayValueString", &nodeAttribute{DisplayValue: nil}, nil, false},
		{"DisplayValueString", &nodeAttribute{DisplayValue: "tests"}, "tests", true},
		{"DisplayValueBool", &nodeAttribute{DisplayValue: nil}, nil, false},
		{"DisplayValueBool", &nodeAttribute{DisplayValue: true}, true, true},
		{"DisplayValueFloat", &nodeAttribute{DisplayValue: nil}, nil, false},
		{"DisplayValueFloat", &nodeAttribute{DisplayValue: 1.0}, 1.0, true},
		{"DisplayValueInt", &nodeAttribute{DisplayValue: nil}, nil, false},
		{"DisplayValueInt", &nodeAttribute{DisplayValue: -100}, int64(-100), true},
		{"DisplayValueUint", &nodeAttribute{DisplayValue: nil}, nil, false},
		{"DisplayValueUint", &nodeAttribute{DisplayValue: 100}, uint64(100), true},
		{"TargetValueString", &nodeAttribute{TargetValue: nil}, nil, false},
		{"TargetValueString", &nodeAttribute{TargetValue: "tests"}, "tests", true},
		{"TargetValueBool", &nodeAttribute{TargetValue: nil}, nil, false},
		{"TargetValueBool", &nodeAttribute{TargetValue: true}, true, true},
		{"TargetValueFloat", &nodeAttribute{TargetValue: nil}, nil, false},
		{"TargetValueFloat", &nodeAttribute{TargetValue: 1.0}, 1.0, true},
		{"TargetValueInt", &nodeAttribute{TargetValue: nil}, nil, false},
		{"TargetValueInt", &nodeAttribute{TargetValue: -100}, int64(-100), true},
		{"TargetValueUint", &nodeAttribute{TargetValue: nil}, nil, false},
		{"TargetValueUint", &nodeAttribute{TargetValue: 100}, uint64(100), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals := reflect.ValueOf(tt.attr).MethodByName(tt.name).Call(nil)
			got := vals[0].Interface()
			ok := vals[1].Interface().(bool)

			if ok != tt.wantOk {
				t.Errorf("nodeAttribute.%v() ok = %v, wantOk = %v", tt.name, ok, tt.wantOk)
				return
			}

			if !ok {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeAttribute.%v() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_nodeAttribute_int64(t *testing.T) {
	tests := []struct {
		name   string
		v      interface{}
		wantI  int64
		wantOk bool
	}{
		{"Int", int(10), 10, true},
		{"Int8", int8(125), 125, true},
		{"Int16", int16(32670), 32670, true},
		{"Int32", int32(2147483640), 2147483640, true},
		{"Int64", int64(9223372036854775800), 9223372036854775800, true},
		{"IntNeg", int(-10), -10, true},
		{"Int8Neg", int8(-125), -125, true},
		{"Int16Neg", int16(-32670), -32670, true},
		{"Int32Neg", int32(-2147483640), -2147483640, true},
		{"Int64Neg", int64(-9223372036854775800), -9223372036854775800, true},
		{"Other", "no", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			na := &nodeAttribute{}
			gotI, gotOk := na.int64(tt.v)
			if gotI != tt.wantI {
				t.Errorf("nodeAttribute.int64() gotI = %v, want %v", gotI, tt.wantI)
			}
			if gotOk != tt.wantOk {
				t.Errorf("nodeAttribute.int64() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_nodeAttribute_uint64(t *testing.T) {
	tests := []struct {
		name   string
		v      interface{}
		wantI  uint64
		wantOk bool
	}{
		{"Uint", uint(10), 10, true},
		{"Uint8", uint8(250), 250, true},
		{"Uint16", uint16(65530), 65530, true},
		{"Uint32", uint32(4294967290), 4294967290, true},
		{"Uint64", uint64(18446744073709551610), 18446744073709551610, true},
		{"Int", int(10), 10, true},
		{"Int8", int8(125), 125, true},
		{"Int16", int16(32670), 32670, true},
		{"Int32", int32(2147483640), 2147483640, true},
		{"Int64", int64(9223372036854775800), 9223372036854775800, true},
		{"Other", "no", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			na := &nodeAttribute{}
			gotI, gotOk := na.uint64(tt.v)
			if gotI != tt.wantI {
				t.Errorf("nodeAttribute.uint64() gotI = %v, want %v", gotI, tt.wantI)
			}
			if gotOk != tt.wantOk {
				t.Errorf("nodeAttribute.uint64() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
