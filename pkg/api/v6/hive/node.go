package hive

import (
	"encoding/json"
	"fmt"
)

type node struct {
	ID            string                    `json:"id,omitempty"`
	Href          string                    `json:"href,omitempty"`
	Name          string                    `json:"name,omitempty"`
	ParentNodeID  string                    `json:"parentNodeId,omitempty"`
	LastSeen      int64                     `json:"lastSeen,omitempty"`
	CreatedOn     int64                     `json:"createdOn,omitempty"`
	UserID        string                    `json:"userId,omitempty"`
	OwnerID       string                    `json:"ownerId,omitempty"`
	HomeID        string                    `json:"homeId,omitempty"`
	Attributes    map[string]*nodeAttribute `json:"attributes,omitempty"`
	Relationships json.RawMessage           `json:"relationships,omitempty"`
}

func (n *node) NodeType() (string, error) {
	if attr, ok := n.Attributes["nodeType"]; ok {
		v, ok := attr.ReportedValueString()
		if !ok {
			return "", &Error{
				Op:      "nodetype",
				Code:    ErrInvalidNodeType,
				Message: fmt.Sprintf("node type attribute not string, %T %v", v, v),
			}
		}

		return v, nil
	}

	return "", &Error{
		Op:      "nodetype",
		Code:    ErrInvalidNodeJSON,
		Message: "node type attribute missing",
	}
}

type nodeAttribute struct {
	ReportedValue      interface{} `json:"reportedValue,omitempty"`
	DisplayValue       interface{} `json:"displayValue,omitempty"`
	TargetValue        interface{} `json:"targetValue,omitempty"`
	ReportReceivedTime int64       `json:"reportReceivedTime,omitempty"` // 1539205419366
	ReportChangedTime  int64       `json:"reportChangedTime,omitempty"`  // 1528575087449
}

func (na *nodeAttribute) int64(v interface{}) (i int64, ok bool) {
	ok = true

	switch v.(type) {
	case int:
		ii := v.(int)
		i = int64(ii)
	case int8:
		ii := v.(int8)
		i = int64(ii)
	case int16:
		ii := v.(int16)
		i = int64(ii)
	case int32:
		ii := v.(int32)
		i = int64(ii)
	case int64:
		ii := v.(int64)
		i = int64(ii)
	default:
		ok = false
	}

	return
}

func (na *nodeAttribute) uint64(v interface{}) (i uint64, ok bool) {
	ok = true

	switch v.(type) {
	case int:
		ii := v.(int)
		i = uint64(ii)
	case uint:
		ii := v.(uint)
		i = uint64(ii)
	case int8:
		ii := v.(int8)
		i = uint64(ii)
	case uint8:
		ii := v.(uint8)
		i = uint64(ii)
	case int16:
		ii := v.(int16)
		i = uint64(ii)
	case uint16:
		ii := v.(uint16)
		i = uint64(ii)
	case int32:
		ii := v.(int32)
		i = uint64(ii)
	case uint32:
		ii := v.(uint32)
		i = uint64(ii)
	case int64:
		ii := v.(int64)
		i = uint64(ii)
	case uint64:
		ii := v.(uint64)
		i = uint64(ii)
	default:
		ok = false
	}

	return
}

func (na *nodeAttribute) ReportedValueString() (s string, ok bool) {
	s, ok = na.ReportedValue.(string)
	return
}

func (na *nodeAttribute) ReportedValueBool() (b bool, ok bool) {
	b, ok = na.ReportedValue.(bool)
	return
}

func (na *nodeAttribute) ReportedValueFloat() (f float64, ok bool) {
	f, ok = na.ReportedValue.(float64)
	return
}

func (na *nodeAttribute) ReportedValueInt() (i int64, ok bool) {
	return na.int64(na.ReportedValue)
}

func (na *nodeAttribute) ReportedValueUint() (i uint64, ok bool) {
	return na.uint64(na.ReportedValue)
}

func (na *nodeAttribute) DisplayValueString() (s string, ok bool) {
	s, ok = na.DisplayValue.(string)
	return
}

func (na *nodeAttribute) DisplayValueBool() (b bool, ok bool) {
	b, ok = na.DisplayValue.(bool)
	return
}

func (na *nodeAttribute) DisplayValueFloat() (f float64, ok bool) {
	f, ok = na.DisplayValue.(float64)
	return
}

func (na *nodeAttribute) DisplayValueInt() (i int64, ok bool) {
	return na.int64(na.DisplayValue)
}

func (na *nodeAttribute) DisplayValueUint() (i uint64, ok bool) {
	return na.uint64(na.DisplayValue)
}

func (na *nodeAttribute) TargetValueString() (s string, ok bool) {
	s, ok = na.TargetValue.(string)
	return
}

func (na *nodeAttribute) TargetValueBool() (b bool, ok bool) {
	b, ok = na.TargetValue.(bool)
	return
}

func (na *nodeAttribute) TargetValueFloat() (f float64, ok bool) {
	f, ok = na.TargetValue.(float64)
	return
}

func (na *nodeAttribute) TargetValueInt() (i int64, ok bool) {
	return na.int64(na.TargetValue)
}

func (na *nodeAttribute) TargetValueUint() (i uint64, ok bool) {
	return na.uint64(na.TargetValue)
}
