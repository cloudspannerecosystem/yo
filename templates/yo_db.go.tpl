// YODB is the common interface for database operations.
type YODB interface {
	YORODB
}

// YORODB is the common interface for database operations.
type YORODB interface {
	ReadRow(ctx context.Context, table string, key spanner.Key, columns []string) (*spanner.Row, error)
	Read(ctx context.Context, table string, keys spanner.KeySet, columns []string) *spanner.RowIterator
	ReadUsingIndex(ctx context.Context, table, index string, keys spanner.KeySet, columns []string) (ri *spanner.RowIterator)
	Query(ctx context.Context, statement spanner.Statement) *spanner.RowIterator
}

// YOLog provides the log func used by generated queries.
var YOLog = func(context.Context, string, ...interface{}) { }

func newError(method, table string, err error) error {
	code := spanner.ErrCode(err)
	return newErrorWithCode(code, method, table, err)
}

func newErrorWithCode(code codes.Code, method, table string, err error) error {
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
	code   codes.Code
}

func (e yoError) Error() string {
	return fmt.Sprintf("yo error in %s(%s): %v", e.method, e.table, e.err)
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
func (e yoError) GRPCStatus() *status.Status {
	var ae *apierror.APIError
	if errors.As(e.err, &ae) {
		return status.Convert(ae)
	}

	return status.New(e.code, e.Error())
}

func (e yoError) Timeout() bool { return e.code == codes.DeadlineExceeded }
func (e yoError) Temporary() bool { return e.code == codes.DeadlineExceeded }
func (e yoError) NotFound() bool { return e.code == codes.NotFound }
