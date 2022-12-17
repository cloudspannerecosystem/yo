// YODB is the common interface for database operations.
type YODB interface {
	ReadRow(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}, table string, key {{ .PackageRegistry.Use presetPackages.goSpanner "Key" }}, columns []string) (*{{ .PackageRegistry.Use presetPackages.goSpanner "Row" }}, error)
	Read(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}, table string, keys {{ .PackageRegistry.Use presetPackages.goSpanner "KeySet" }}, columns []string) *{{ .PackageRegistry.Use presetPackages.goSpanner "RowIterator" }}
	ReadUsingIndex(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}, table, index string, keys {{ .PackageRegistry.Use presetPackages.goSpanner "KeySet" }}, columns []string) (ri *{{ .PackageRegistry.Use presetPackages.goSpanner "RowIterator" }})
	Query(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}, statement {{ .PackageRegistry.Use presetPackages.goSpanner "Statement" }}) *{{ .PackageRegistry.Use presetPackages.goSpanner "RowIterator" }}
}

// YOLog provides the log func used by generated queries.
var YOLog = func({{ .PackageRegistry.Use presetPackages.context "Context" }}, string, ...interface{}) { }

func newError(method, table string, err error) error {
	code := {{ .PackageRegistry.Use presetPackages.goSpanner "ErrCode" }}(err)
	return newErrorWithCode(code, method, table, err)
}

func newErrorWithCode(code {{ .PackageRegistry.Use presetPackages.gRPCCodes "Code" }}, method, table string, err error) error {
	return &yoError{
		method: method,
		table:  table,
		err:    err,
		code:   code,
	}
}

type yoError struct {
	err    error
	method string
	table  string
	code   {{ .PackageRegistry.Use presetPackages.gRPCCodes "Code" }}
}

func (e yoError) Error() string {
	return {{ .PackageRegistry.Use presetPackages.fmt "Sprintf" }}("yo error in %s(%s): %v", e.method, e.table, e.err)
}

func (e yoError) Unwrap() error {
	return e.err
}

func (e yoError) DBTableName() string {
	return e.table
}

// GRPCStatus implements a conversion to a gRPC status using `status.Convert(error)`.
// If the error is originated from the Spanner library, this returns a gRPC status of
// the original error. It may contain details of the status such as RetryInfo.
func (e yoError) GRPCStatus() *{{ .PackageRegistry.Use presetPackages.gRPCStatus "Status" }} {
	var ae *{{ .PackageRegistry.Use presetPackages.googleApisGaxGoV2ApiError "APIError" }}
	if {{ .PackageRegistry.Use presetPackages.errors "As" }}(e.err, &ae) {
		return {{ .PackageRegistry.Use presetPackages.gRPCStatus "Convert" }}(ae)
	}

	return {{ .PackageRegistry.Use presetPackages.gRPCStatus "New" }}(e.code, e.Error())
}

func (e yoError) Timeout() bool { return e.code == {{ .PackageRegistry.Use presetPackages.gRPCCodes "DeadlineExceeded" }} }
func (e yoError) Temporary() bool { return e.code == {{ .PackageRegistry.Use presetPackages.gRPCCodes "DeadlineExceeded" }} }
func (e yoError) NotFound() bool { return e.code == {{ .PackageRegistry.Use presetPackages.gRPCCodes "NotFound" }} }

// yoEncode encodes primitive types that spanner library does not support into spanner types before
// passing to spanner functions. Supported primitive types and user defined types that implement
// spanner.Encoder interface are handled in encoding phase inside Spanner library.
func yoEncode(v interface{}) interface{} {
	switch vv := v.(type) {
	case int8:
		return int64(vv)
	case uint8:
		return int64(vv)
	case int16:
		return int64(vv)
	case uint16:
		return int64(vv)
	case int32:
		return int64(vv)
	case uint32:
		return int64(vv)
	case uint64:
		return int64(vv)
	default:
		return v
	}
}

// yoDecode wraps primitive types that spanner library does not support to decode from spanner types
// by yoPrimitiveDecoder before passing to Spanner functions. Supported primitive types and
// user defined types that implement spanner.Decoder interface are handled in decoding phase inside
// Spanner library.
func yoDecode(ptr interface{}) interface{} {
	switch ptr.(type) {
	case *int8, *uint8, *int16, *uint16, *int32, *uint32, *uint64:
		return &yoPrimitiveDecoder{val: ptr}
	default:
		return ptr
	}
}

type yoPrimitiveDecoder struct {
	val interface{}
}

func (y *yoPrimitiveDecoder) DecodeSpanner(val interface{}) error {
	strVal, ok := val.(string)
	if !ok {
		return {{ .PackageRegistry.Use presetPackages.goSpanner "ToSpannerError" }}({{ .PackageRegistry.Use presetPackages.gRPCStatus "Errorf" }}({{ .PackageRegistry.Use presetPackages.gRPCCodes "FailedPrecondition" }}, "failed to decode customField: %T(%v)", val, val))
	}

	intVal, err := {{ .PackageRegistry.Use presetPackages.strconv "ParseInt" }}(strVal, 10, 64)
	if err != nil {
		return {{ .PackageRegistry.Use presetPackages.goSpanner "ToSpannerError" }}({{ .PackageRegistry.Use presetPackages.gRPCStatus "Errorf" }}({{ .PackageRegistry.Use presetPackages.gRPCCodes "FailedPrecondition" }}, "%v wasn't correctly encoded: <%v>", val, err))
	}

	switch vv := y.val.(type) {
	case *int8:
		*vv = int8(intVal)
	case *uint8:
		*vv = uint8(intVal)
	case *int16:
		*vv = int16(intVal)
	case *uint16:
		*vv = uint16(intVal)
	case *int32:
		*vv = int32(intVal)
	case *uint32:
		*vv = uint32(intVal)
	case *uint64:
		*vv = uint64(intVal)
	default:
		return {{ .PackageRegistry.Use presetPackages.gRPCStatus "Errorf" }}({{ .PackageRegistry.Use presetPackages.gRPCCodes "Internal" }}, "unexpected type for yoPrimitiveDecoder: %T", y.val)
	}

	return nil
}
