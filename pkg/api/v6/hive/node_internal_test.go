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
