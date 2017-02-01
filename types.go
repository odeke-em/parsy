package parsy

import (
	"fmt"
	"strconv"
)

type Type uint

const (
	TInterface Type = 1 + iota
	TBool
	TInt
	TInt64
	TUint
	TUint64
	TFloat32
	TFloat64
	TString
)

type ParseFn func(string) (interface{}, error)

func (typ Type) Parser() (ParseFn, error) {
	switch typ {
	case TBool:
		return parseBool, nil
	case TInt:
		return parseInt, nil
	case TString:
		return parseString, nil
	case TUint:
		return parseUint, nil
	case TUint64:
		return parseUint64, nil
	case TFloat32:
		return parseFloat32, nil
	case TFloat64:
		return parseFloat64, nil
	default:
		return nil, fmt.Errorf("unknown type: %d", typ)
	}
}

func parseBool(s string) (interface{}, error) {
	return strconv.ParseBool(s)
}

func parseString(s string) (interface{}, error) {
	return s, nil
}

func parseFloat64(s string) (interface{}, error) {
	return strconv.ParseFloat(s, 64)
}

func parseFloat32(s string) (interface{}, error) {
	f64, err := strconv.ParseFloat(s, 32)
	return float32(f64), err
}

func parseUint(s string) (interface{}, error) {
	u64, err := strconv.ParseUint(s, 10, 32)
	return uint(u64), err
}

func parseInt(s string) (interface{}, error) {
	i64, err := strconv.ParseInt(s, 10, 32)
	return int(i64), err
}

func parseInt64(s string) (interface{}, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseInt32(s string) (interface{}, error) {
	i64, err := strconv.ParseInt(s, 10, 32)
	return int32(i64), err
}

func parseUint64(s string) (interface{}, error) {
	u64, err := strconv.ParseUint(s, 10, 64)
	return u64, err
}

func parseUint32(s string) (interface{}, error) {
	u64, err := strconv.ParseUint(s, 10, 64)
	return uint32(u64), err
}
