package response

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strconv"

	"github.com/d1vbyz3r0/typed/common/meta"
	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser/calls"
	"github.com/d1vbyz3r0/typed/internal/parser/response/codes"
	"github.com/d1vbyz3r0/typed/internal/parser/response/mime"
	"github.com/labstack/echo/v4"
)

const (
	jsonContextFunc       = "JSON"
	jsonPrettyContextFunc = "JSONPretty"
	jsonBlobContextFunc   = "JSONBlob"
	xmlContextFunc        = "XML"
	xmlPrettyContextFunc  = "XMLPretty"
	xmlBlobContextFunc    = "XMLBlob"
	stringContextFunc     = "String"
	blobContextFunc       = "Blob"
	redirectContextFunc   = "Redirect"
	noContentContextFunc  = "NoContent"
	streamContextFunc     = "Stream"
)

type ContextResponseType struct {
	funcName string
	call     *ast.CallExpr
	codes    *codes.Resolver
	mime     *mime.Resolver
	types    *types.Info
}

var supportedFunctions = []string{
	jsonContextFunc, jsonPrettyContextFunc, jsonBlobContextFunc,
	xmlContextFunc, xmlPrettyContextFunc, xmlBlobContextFunc,
	stringContextFunc,
	blobContextFunc,
	redirectContextFunc,
	noContentContextFunc,
	streamContextFunc,
}

var (
	rawBodyFuncs = []string{jsonBlobContextFunc, xmlBlobContextFunc, streamContextFunc}
	noBodyFuncs  = []string{redirectContextFunc, noContentContextFunc}
)

func newContextResponseType(
	call *ast.CallExpr,
	cr *codes.Resolver,
	mr *mime.Resolver,
	typesInfo *types.Info,
) (t ContextResponseType, supported bool) {
	if !calls.IsEchoContextMethodCall(call) {
		return ContextResponseType{}, false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ContextResponseType{}, false
	}

	t = ContextResponseType{
		funcName: sel.Sel.Name,
		call:     call,
		codes:    cr,
		mime:     mr,
		types:    typesInfo,
	}

	supported = slices.Contains(supportedFunctions, t.funcName)
	return t, supported
}

// ContentType returns content type retrieved from func usage context. It's empty for Redirect and NoContent
func (t ContextResponseType) ContentType() (string, error) {
	switch t.funcName {
	case jsonContextFunc, jsonPrettyContextFunc, jsonBlobContextFunc:
		return echo.MIMEApplicationJSON, nil

	case xmlContextFunc, xmlPrettyContextFunc, xmlBlobContextFunc:
		return echo.MIMEApplicationXML, nil

	case stringContextFunc:
		return echo.MIMETextPlain, nil

	case blobContextFunc:
		contentTypeArg := t.call.Args[1]
		contentType, err := t.getContentTypeFromArg(contentTypeArg)
		if err != nil {
			return "", fmt.Errorf("get content type from arg: %w", err)
		}
		return contentType, nil

	case redirectContextFunc:
		return "", nil

	case noContentContextFunc:
		return "", nil

	case streamContextFunc:
		contentTypeArg := t.call.Args[1]
		contentType, err := t.getContentTypeFromArg(contentTypeArg)
		if err != nil {
			return "", fmt.Errorf("get content type from arg: %w", err)
		}
		return contentType, nil

	default:
		return "", fmt.Errorf("unexpected function name %s", t.funcName)
	}
}

func (t ContextResponseType) StatusCode() (int, error) {
	code, err := t.codes.Resolve(t.call.Args[0])
	if err != nil {
		return 0, fmt.Errorf("resolve status code for %s: %w", t.funcName, err)
	}

	return code, nil
}

func (t ContextResponseType) getContentTypeFromArg(arg ast.Expr) (string, error) {
	switch contentTypeArg := arg.(type) {
	case *ast.BasicLit:
		if contentTypeArg.Kind != token.STRING {
			return "", fmt.Errorf("expected string literal, got %s", contentTypeArg.Kind)
		}

		contentType, err := strconv.Unquote(contentTypeArg.Value)
		if err != nil {
			return "", fmt.Errorf("unquote content type: %w", err)
		}

		return contentType, nil

	case *ast.SelectorExpr:
		contentType, err := t.mime.Resolve(contentTypeArg)
		if err != nil {
			return "", fmt.Errorf("resolve content type: %w", err)
		}

		return contentType, nil

	default:
		return "", fmt.Errorf("expected BasicLit or SelectorExpr, got %T", contentTypeArg)
	}
}

func (t ContextResponseType) TypeName() (string, error) {
	if slices.Contains(rawBodyFuncs, t.funcName) || slices.Contains(noBodyFuncs, t.funcName) {
		return "", nil
	}

	name, err := meta.GetTypeName(t.types.TypeOf(t.call.Args[1]))
	if err != nil {
		return "", fmt.Errorf("get type name for %s: %w", t.funcName, err)
	}

	return name, nil
}

func (t ContextResponseType) TypePkgPath() (string, error) {
	if slices.Contains(rawBodyFuncs, t.funcName) || slices.Contains(noBodyFuncs, t.funcName) {
		return "", nil
	}

	argType := t.types.TypeOf(t.call.Args[1])
	if typing.IsAnyType(argType) {
		return "", nil
	}

	if typing.IsBasicType(argType) {
		return "", nil
	}

	if typing.IsMap(argType) || typing.IsSlice(argType) {
		elemType, ok := typing.GetUnderlyingElemType(argType)
		if !ok {
			return "", fmt.Errorf("failed to get underlying elem type for %s", argType)
		}

		if typing.IsAnyType(elemType) {
			return "", nil
		}

		if typing.IsBasicType(elemType) {
			return "", nil
		}

		path, err := meta.GetPkgPath(elemType)
		if err != nil {
			return "", fmt.Errorf("get package path for %s: %w", argType, err)
		}

		return path, nil
	}

	path, err := meta.GetPkgPath(argType)
	if err != nil {
		return "", fmt.Errorf("get pkg path for %s: %w", argType, err)
	}

	return path, nil
}
