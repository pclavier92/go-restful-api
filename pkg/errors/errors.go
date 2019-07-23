package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pclavier92/go-restful-api/pkg/logs"
)

// Pkg will return an error handler for a package
func Pkg(pkg string, l logs.Printer) Package {
	return Package{l, pkg}
}

// Pkger is a type which returns struct level handler for errors
type Pkger interface {
	Struct(s string) Struct
}

// Package wraps errors of a whole package
type Package struct {
	Log  logs.Printer
	name string
}

// Struct will return an error handler for a struct
func (p Package) Struct(strct string) Struct {
	return Struct{p, strct}
}

// Structer is a type which returns a function level hanlder for errors
type Structer interface {
	Fn(s string) Function
}

// Struct will wrap errors for a specific struct
type Struct struct {
	pkg  Package
	name string
}

// Fn will give you something to wrap the error in your functions
func (s Struct) Fn(fn string) Function {
	return Function{s, make(map[string]interface{}), fn}
}

// Creator is a type which can create errors
type Creator interface {
	New(ctx string) *Chain
	Wrap(e error, ctx string) *Chain
	JSON(e error, ctx string) *Chain
	NotFound() *Chain
	DB(e error) *Chain
	Invalid(ctx string) *Chain
	UK(e error) *Chain
}

// Function wraps errors inside a single function
type Function struct {
	strct Struct
	tags  map[string]interface{}
	name  string
}

func (f Function) unsafeWrap(e error, ctx string, external string) *Chain {
	return &Chain{e, external, ctx, f, f.strct, f.strct.pkg}
}

// New will return a new error chain
func (f Function) New(ctx string) error {
	return f.unsafeWrap(errors.New(ctx), ctx, "")
}

// Tag will add a tag to a function error handler
func (f Function) Tag(t string, v interface{}) Function {
	f.tags[t] = v
	return f
}

// Wrap will wrap an error with the function information
func (f Function) Wrap(e error, ctx string, args ...interface{}) error {
	if e == nil {
		return nil
	}
	return f.unsafeWrap(e, fmt.Sprintf(ctx, args...), "")
}

// JSON will wrap an error for display in the API blaming a JSON received
func (f Function) JSON(e error, ctx string) error {
	return f.unsafeWrap(e, ctx, "invalid json")
}

// NotFound will wrap an error saying the resource was not found
func (f Function) NotFound() error {
	return f.unsafeWrap(f.New("not found"), "", "resource not found")
}

// UK will tell the user the problem is unknown
func (f Function) UK(e error) error {
	if e == nil {
		e = f.New("unkown error")
	}
	return f.unsafeWrap(e, "", "unknown error")
}

// DB will wrap an error and add to the context that there was a problem in the database
func (f Function) DB(e error) error {
	return f.unsafeWrap(e, "", "problem in database")
}

// Invalid will create a new error chain and say context is the
func (f Function) Invalid(ctx string) error {
	return f.unsafeWrap(errors.New(ctx), ctx, "invalid data sent")
}

type Chain struct {
	previous error
	External string
	ctx      string
	fn       Function
	strct    Struct
	Pkg      Package
}

func (w Chain) formatTags(tags map[string]interface{}) string {
	var fmtTags string
	for t, v := range w.fn.tags {
		if fmtTags == "" {
			fmtTags += fmt.Sprintf("%s->%v", t, v)
		} else {
			fmtTags += fmt.Sprintf(", %s->%v", t, v)
		}
	}
	return fmtTags
}

func (w Chain) Error() string {
	allTags := w.fn.tags
	stack := fmt.Sprintf("%s.%s.%s", w.Pkg.name, w.fn.strct.name, w.fn.name)
	ctx := w.ctx
	original := w.previous
	prev, ok := w.previous.(*Chain)
	for ok {
		stack += fmt.Sprintf(" <- %s.%s.%s", prev.Pkg.name, prev.strct.name, prev.fn.name)
		ctx += fmt.Sprintf(" <- %s", prev.ctx)
		for t, v := range prev.fn.tags {
			allTags[t] = v
		}
		if _, ok = prev.previous.(*Chain); ok {
			prev = prev.previous.(*Chain)
		} else {
			original = prev.previous
		}
	}
	tags := w.formatTags(allTags)
	origin := "nil"
	if original != nil {
		origin = original.Error()
	}
	return fmt.Sprintf(
		"ERROR %s | CONTEXT %s | TAGS: %s | STACK: %s",
		origin, ctx, tags, stack)
}

func OriginalError(err error) string {
	original := strings.TrimPrefix(err.Error(), "ERROR ")
	original = strings.SplitAfter(original, " | CONTEXT ")[0]
	return strings.TrimSuffix(original, " | CONTEXT ")
}

func New(msg string) error {
	return errors.New(msg)
}
