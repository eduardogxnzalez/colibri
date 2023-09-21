package colibri

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

var (
	// ErrMustBeConvBool is returned when the value is not convertible to bool.
	ErrMustBeConvBool = errors.New("must be a bool, string or number")

	// ErrMustBeConvDuration is returned when the value is not convertible to time.Duration.
	ErrMustBeConvDuration = errors.New("must be a string or number")

	// ErrMustBeString is returned when the value must be a string.
	ErrMustBeString = errors.New("must be a string")

	// ErrInvalidHeader is returned when the header is invalid.
	ErrInvalidHeader = errors.New("invalid header")
)

// ConvFunc processes the value based on the key.
type ConvFunc func(key string, rawValue any) (any, error)

// DefaultConvFunc ConvFunc used by default by the NewRules function.
func DefaultConvFunc(key string, rawValue any) (any, error) {
	switch key {
	case KeyURL, KeyProxy:
		return ToURL(rawValue)

	case KeyIgnoreRobotsTxt, KeyFollow, KeyUseCookies, KeyAll:
		return toBool(rawValue)

	case KeyDelay, KeyTimeout:
		return toDuration(rawValue)

	case KeyHeader:
		return toHeader(rawValue)

	case KeySelectors:
		return newSelectors(rawValue, DefaultConvFunc)
	}
	return rawValue, nil
}

// ToURL converts a value to a *url.URL.
func ToURL(value any) (*url.URL, error) {
	rawURL, ok := value.(string)
	if ok {
		return url.Parse(rawURL)
	}
	return nil, ErrMustBeString
}

// toBool converts a value to a boolean.
func toBool(value any) (bool, error) {
	if value == nil {
		return false, nil
	}

	switch rValue := reflect.ValueOf(value); rValue.Kind() {
	case reflect.String:
		return strconv.ParseBool(value.(string))

	case reflect.Bool:
		return value.(bool), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rValue.Int() != 0, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rValue.Uint() != 0, nil

	case reflect.Float32, reflect.Float64:
		return rValue.Float() != 0, nil
	}

	return false, ErrMustBeConvBool
}

// toDuration converts a value to a time.Duration.
func toDuration(value any) (time.Duration, error) {
	if value == nil {
		return 0, nil
	}

	switch rValue := reflect.ValueOf(value); rValue.Kind() {
	case reflect.String:
		return time.ParseDuration(value.(string))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return time.Duration(rValue.Int()) * time.Millisecond, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return time.Duration(rValue.Uint()) * time.Millisecond, nil

	case reflect.Float32, reflect.Float64:
		return time.Duration(rValue.Float()*1000000) * time.Nanosecond, nil
	}

	return 0, ErrMustBeConvDuration
}

// toHeader converts a value to a http.Header.
func toHeader(value any) (http.Header, error) {
	if value == nil {
		return http.Header{}, nil
	}

	rValue := reflect.ValueOf(value)
	if rValue.Kind() != reflect.Map {
		return http.Header{}, ErrInvalidHeader
	}

	var (
		header = http.Header{}
		iter   = rValue.MapRange()
	)
	for iter.Next() {
		k := iter.Key()
		if k.Kind() != reflect.String {
			return header, ErrInvalidHeader
		}

		v := iter.Value().Interface()
		switch value := v.(type) {
		case string:
			header.Set(k.String(), value)
			continue

		case []string:
			key := k.String()
			for _, e := range value {
				header.Add(key, e)
			}
			continue
		}

		return header, ErrInvalidHeader
	}

	return header, nil
}
