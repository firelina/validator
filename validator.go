package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrNotStruct                   = errors.New("wrong argument given, should be a struct")
	ErrInvalidValidatorSyntax      = errors.New("invalid validator syntax")
	ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")
	ErrLenValidationFailed         = errors.New("len validation failed")
	ErrInValidationFailed          = errors.New("in validation failed")
	ErrMaxValidationFailed         = errors.New("max validation failed")
	ErrMinValidationFailed         = errors.New("min validation failed")
)

type Valid struct {
	errors []error
}

func NewValid() *Valid {
	return &Valid{errors: make([]error, 0)}
}

func (s *Valid) ValidLen(val string, comp int, name string) {
	if len(val) != comp {
		s.errors = append(s.errors, NewValidationError(ErrLenValidationFailed, name))
	}
}

func (s *Valid) ValidLenSyntax(param, name string) (int, bool) {
	length, err1 := strconv.Atoi(param)
	if err1 != nil || length < 0 {
		s.errors = append(s.errors, NewValidationError(ErrInvalidValidatorSyntax, name))
		return 0, false
	}
	return length, true
}

func (s *Valid) ValidIntSyntax(param, name string) (int, bool) {
	minn, err1 := strconv.Atoi(param)
	if err1 != nil {
		s.errors = append(s.errors, NewValidationError(ErrInvalidValidatorSyntax, name))
		return 0, false
	}
	return minn, true
}

func (s *Valid) ValidInIntSyntax(param, name string) ([]int, bool) {
	if len(param) == 0 {
		s.errors = append(s.errors, NewValidationError(ErrInvalidValidatorSyntax, name))
		return nil, false
	}
	ins := strings.Split(param, ",")
	intSlice := make([]int, len(ins))
	for j, str := range ins {
		num, _ := strconv.Atoi(str)
		intSlice[j] = num
	}
	return intSlice, true
}

func (s *Valid) ValidInStringSyntax(param, name string) ([]string, bool) {
	if len(param) == 0 {
		s.errors = append(s.errors, NewValidationError(ErrInvalidValidatorSyntax, name))
		return nil, false
	}
	ins := strings.Split(param, ",")
	return ins, true
}

func (s *Valid) ValidMin(val, comp int, name string) {
	if val < comp {
		s.errors = append(s.errors, NewValidationError(ErrMinValidationFailed, name))
	}
}

func (s *Valid) ValidMax(val, comp int, name string) {
	if val > comp {
		s.errors = append(s.errors, NewValidationError(ErrMaxValidationFailed, name))
	}
}

func (s *Valid) ValidInInt(val int, comp []int, name string) {
	if !Contains(comp, val) {
		s.errors = append(s.errors, NewValidationError(ErrInValidationFailed, name))
	}
}

func (s *Valid) ValidInString(val string, comp []string, name string) {
	if !Contains(comp, val) {
		s.errors = append(s.errors, NewValidationError(ErrInValidationFailed, name))
	}
}

type ValidationError struct {
	field string
	err   error
}

func NewValidationError(err error, field string) error {
	return &ValidationError{
		field: field,
		err:   err,
	}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.field, e.err)
}

func (e *ValidationError) Unwrap() error {
	return e.err
}

func Contains[T comparable](t []T, needle T) bool {
	for _, v := range t {
		if v == needle {
			return true
		}
	}
	return false
}

func (s *Valid) validateMax(val, fieldName string, fieldValue reflect.Value) {
	minn, ok := s.ValidIntSyntax(val, fieldName)
	if !ok {
		return
	}
	if fieldValue.Kind() == reflect.String {
		s.ValidMax(fieldValue.Len(), minn, fieldName)
	}
	if fieldValue.Kind() == reflect.Int {
		s.ValidMax(int(fieldValue.Int()), minn, fieldName)
	}
	if fieldValue.Kind() == reflect.Slice {
		for v := 0; v < fieldValue.Len(); v++ {
			s.validateMax(val, fieldName, fieldValue.Index(v))
		}
	}
}

func (s *Valid) validateMin(val, fieldName string, fieldValue reflect.Value) {
	minn, ok := s.ValidIntSyntax(val, fieldName)
	if !ok {
		return
	}
	if fieldValue.Kind() == reflect.String {
		s.ValidMin(fieldValue.Len(), minn, fieldName)
	}
	if fieldValue.Kind() == reflect.Int {
		s.ValidMin(int(fieldValue.Int()), minn, fieldName)
	}
	if fieldValue.Kind() == reflect.Slice {
		for v := 0; v < fieldValue.Len(); v++ {
			s.validateMin(val, fieldName, fieldValue.Index(v))
		}
	}
}

func (s *Valid) validateLen(val, fieldName string, fieldValue reflect.Value) {
	p, ok := s.ValidLenSyntax(val, fieldName)
	if !ok {
		return
	}
	if fieldValue.Kind() == reflect.String {
		s.ValidLen(fieldValue.String(), p, fieldName)
	}
	if fieldValue.Kind() == reflect.Slice {
		for v := 0; v < fieldValue.Len(); v++ {
			s.validateLen(val, fieldName, fieldValue.Index(v))
		}
	}
}

func (s *Valid) validateIn(val, fieldName string, fieldValue reflect.Value) {
	if fieldValue.Kind() == reflect.String {
		ins, ok := s.ValidInStringSyntax(val, fieldName)
		if !ok {
			return
		}
		s.ValidInString(fieldValue.String(), ins, fieldName)
	}
	if fieldValue.Kind() == reflect.Int {
		ins, ok := s.ValidInIntSyntax(val, fieldName)
		if !ok {
			return
		}
		s.ValidInInt(int(fieldValue.Int()), ins, fieldName)
	}
	if fieldValue.Kind() == reflect.Slice {
		for v := 0; v < fieldValue.Len(); v++ {
			s.validateIn(val, fieldName, fieldValue.Index(v))
		}
	}
}

func Validate(s any) error {
	// TODO implement me
	valueOfS := reflect.ValueOf(s)
	if valueOfS.Type().Kind() == reflect.Ptr {
		valueOfS = valueOfS.Elem()
	}
	if valueOfS.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	valid := NewValid()
	for i := 0; i < valueOfS.NumField(); i++ {
		fieldValue := valueOfS.Field(i)
		fieldName := valueOfS.Type().Field(i).Name
		tag := valueOfS.Type().Field(i).Tag.Get("validate")
		if tag == "" {
			continue
		}
		if !valueOfS.Type().Field(i).IsExported() {
			return ErrValidateForUnexportedFields
		}
		tagParts := strings.Split(tag, ":")
		val := tagParts[1]
		switch tagParts[0] {
		case "len":
			valid.validateLen(val, fieldName, fieldValue)
		case "in":
			valid.validateIn(val, fieldName, fieldValue)
		case "min":
			valid.validateMin(val, fieldName, fieldValue)
		case "max":
			valid.validateMax(val, fieldName, fieldValue)
		}
		fmt.Println(fieldName, fieldValue.Interface())
	}
	return errors.Join(valid.errors...)
}
