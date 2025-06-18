package generator

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type QueryParam struct {
	Name     string
	Required bool   // false for c.QueryParam calls
	Type     string // Inferred from usage, for now always string, except for struct fields
}

type HandlerInfo struct {
	Name        string        // Function name
	Package     string        // Package path relative to base
	IsWrapper   bool          // If returns echo.HandlerFunc
	TypesInfo   *types.Info   // Store type information for later analysis TODO: unused, delete?
	File        *ast.File     // AST file containing the handler TODO: unused, delete?
	Node        *ast.FuncDecl // Function declaration node TODO: unused, delete?
	RequestDTO  *DTOInfo
	Responses   map[int]*ResponseInfo // Map status code to response type
	QueryParams []*QueryParam
	Doc         string
}

type FormField struct {
	Name     string
	Type     string // Type can be object (for files) or any basic type ???
	Required bool
	IsFile   bool
	IsArray  bool
}

type DTOInfo struct {
	Type        types.Type
	TypeName    string
	Package     string
	ContentType string
	FormFields  []*FormField
	IsForm      bool
}

type ResponseInfo struct {
	StatusCode  int
	Type        types.Type
	TypeName    string
	IsArray     bool
	IsMap       bool
	KeyType     string
	ValueType   string
	Package     string
	ContentType string
	NoContent   bool
}

type HandlerAnalyzer struct {
	pkgs         []PkgInfo
	handlers     map[string]*HandlerInfo
	codeResolver *statusCodeResolver
}

type PkgInfo struct {
	Path      string
	Recursive bool
}

func NewHandlerAnalyzer(pkgs []PkgInfo) *HandlerAnalyzer {
	return &HandlerAnalyzer{
		pkgs:         pkgs,
		handlers:     make(map[string]*HandlerInfo),
		codeResolver: newStatusCodeResolver(),
	}
}

func (ha *HandlerAnalyzer) DiscoverHandlers() error {
	patterns, err := ha.buildPatterns()
	if err != nil {
		return fmt.Errorf("build patterns: %v", err)
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	logger.Debug("Loaded packages", "count", len(pkgs), "patterns", patterns)

	// Process each package under the base handlers directory
	for _, pkg := range pkgs {
		logger.Debug("Analyzing package", "path", pkg.Types.Path())
		err := ha.analyzePackage(pkg)
		if err != nil {
			return fmt.Errorf("analyze package: %w", err)
		}

		if pkg.Errors != nil && len(pkg.Errors) > 0 {
			err := fmt.Errorf("%w", pkg.Errors[0])
			for i := 1; i < len(pkg.Errors); i++ {
				err = fmt.Errorf("%w: %w", err, pkg.Errors[i])
			}

			return fmt.Errorf("analyze package %s: %w", pkg.String(), err)
		}
	}

	logger.Info("Discovering finished", "handlers_count", len(ha.handlers))

	return nil
}

func (ha *HandlerAnalyzer) Handlers() map[string]*HandlerInfo {
	return ha.handlers
}

func (ha *HandlerAnalyzer) analyzePackage(pkg *packages.Package) error {
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				// Check if function returns echo.HandlerFunc or is echo.HandlerFunc
				if ha.isHandlerFunction(funcDecl, pkg.TypesInfo) {
					responses := ha.analyzeResponses(funcDecl, pkg.TypesInfo)
					logger.Debug("Found responses", "responses", responses)

					requestDTO := ha.analyzeRequestDTO(funcDecl, pkg.TypesInfo)
					if requestDTO != nil {
						logger.Debug("Found request dto", "dto", requestDTO)
					}

					queryParams := ha.analyzeQueryParams(funcDecl, pkg.TypesInfo)
					if len(queryParams) > 0 {
						logger.Debug("Found query params", "params", queryParams)
					}

					doc := ha.extractDocumentation(funcDecl)

					handlerInfo := &HandlerInfo{
						Name:        funcDecl.Name.String(),
						Package:     pkg.Types.Path(),
						IsWrapper:   ha.isWrapperFunction(funcDecl, pkg.TypesInfo),
						TypesInfo:   pkg.TypesInfo,
						File:        file,
						Node:        funcDecl,
						Responses:   responses,
						RequestDTO:  requestDTO,
						QueryParams: queryParams,
						Doc:         doc,
					}
					// Use package path + name as key
					key := pkg.Types.Name() + "." + handlerInfo.Name
					ha.handlers[key] = handlerInfo

					logger.Debug("Found handler", "key", key, "info", *handlerInfo)
				}
			}
			return true
		})
	}
	return nil
}

func (ha *HandlerAnalyzer) isHandlerFunction(funcDecl *ast.FuncDecl, info *types.Info) bool {
	// Check if function is echo.HandlerFunc directly
	if ha.isEchoHandler(funcDecl, info) {
		return true
	}

	// Check if function returns echo.HandlerFunc
	if ha.isWrapperFunction(funcDecl, info) {
		return true
	}

	return false
}

func (ha *HandlerAnalyzer) isEchoHandler(funcDecl *ast.FuncDecl, info *types.Info) bool {
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) != 1 {
		return false
	}

	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) != 1 {
		return false
	}

	// Check if param is echo.Context using selector expression
	paramType := funcDecl.Type.Params.List[0].Type
	if sel, ok := paramType.(*ast.SelectorExpr); ok {
		if x, ok := sel.X.(*ast.Ident); ok {
			if x.Name != "echo" || sel.Sel.Name != "Context" {
				return false
			}
		}
	}

	// Check if return type is error
	returnType := info.TypeOf(funcDecl.Type.Results.List[0].Type)
	return returnType.String() == "error"
}

func (ha *HandlerAnalyzer) isWrapperFunction(funcDecl *ast.FuncDecl, info *types.Info) bool {
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) != 1 {
		return false
	}

	result := funcDecl.Type.Results.List[0].Type
	if sel, ok := result.(*ast.SelectorExpr); ok {
		if x, ok := sel.X.(*ast.Ident); ok {
			// Check if it's from echo package and type is HandlerFunc
			return x.Name == "echo" && sel.Sel.Name == "HandlerFunc"
		}
	}

	return false
}

// TODO: for now it can process only json responses, in future it can be extended
func (ha *HandlerAnalyzer) analyzeResponses(funcDecl *ast.FuncDecl, info *types.Info) map[int]*ResponseInfo {
	responses := make(map[int]*ResponseInfo)

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				_func := sel.Sel.Name

				if _func == "JSON" {
					if len(call.Args) == 2 {
						statusCode := ha.codeResolver.resolve(call.Args[0], info)
						respInfo := &ResponseInfo{
							StatusCode:  statusCode,
							ContentType: "application/json",
						}

						//responseType := info.TypeOf(call.Args[1])
						//switch rt := responseType.(type) {
						//// TODO: process maps as objects
						//case *types.Map:
						//	log.Printf("Got map")
						//	respInfo.Type = rt
						//
						//case *types.Slice:
						//	respInfo.IsArray = true
						//	responseType = rt.Elem()
						//
						//	//default:
						//	//	logger.Info("got into default", "type", rt.Underlying().String())
						//	//	responseType = rt.Underlying()
						//}
						//
						//respInfo.Type = responseType
						//respInfo.TypeName = ha.extractTypeName(responseType)
						//respInfo.Package = ha.extractPackage(responseType)
						//responses[statusCode] = respInfo
						responseType := info.TypeOf(call.Args[1])
						switch rt := responseType.(type) {
						case *types.Named:
							if rt.Obj().Name() == "Map" && rt.Obj().Pkg().Path() == "github.com/labstack/echo/v4" {
								respInfo.Type = rt
								respInfo.TypeName = "echo.Map"
								respInfo.Package = rt.Obj().Pkg().Path()
								respInfo.IsMap = true
								respInfo.KeyType = "string"
								respInfo.ValueType = "any"
							} else {
								respInfo.Type = rt
								respInfo.TypeName = ha.extractTypeName(rt)
								respInfo.Package = ha.extractPackage(rt)
							}

						case *types.Map:
							respInfo.Type = rt
							respInfo.TypeName = fmt.Sprintf("map[%s]%s", rt.Key().String(), rt.Elem().String())
							respInfo.Package = "builtin"
							respInfo.IsMap = true
							respInfo.KeyType = rt.Key().String()
							respInfo.ValueType = rt.Elem().String()

						case *types.Slice:
							respInfo.IsArray = true
							responseType = rt.Elem()
							respInfo.Type = responseType
							respInfo.TypeName = ha.extractTypeName(responseType)
							respInfo.Package = ha.extractPackage(responseType)

						default:
							respInfo.Type = responseType
							respInfo.TypeName = ha.extractTypeName(responseType)
							respInfo.Package = ha.extractPackage(responseType)
						}

						responses[statusCode] = respInfo
						logger.Debug("Found response", "status_code", statusCode, "info", *respInfo)
					}
				} else if _func == "NoContent" {
					if len(call.Args) == 1 {
						statusCode := ha.codeResolver.resolve(call.Args[0], info)
						respInfo := &ResponseInfo{
							StatusCode: statusCode,
							NoContent:  true,
						}
						responses[statusCode] = respInfo

						logger.Debug("Found response", "status_code", statusCode, "info", *respInfo)
					}
				}
			}
		}
		return true
	})

	return responses
}

// TODO: for now it can process only json requests, in future it can be extended
// analyzeRequestDTO will extract dto name and query params, bound with c.Bind()
func (ha *HandlerAnalyzer) analyzeRequestDTO(funcDecl *ast.FuncDecl, info *types.Info) *DTOInfo {
	var dtoInfo *DTOInfo // Request can contain only single body DTO / Form

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Bind" {
					if len(call.Args) > 0 {
						argType := info.TypeOf(call.Args[0])
						if ha.isJSONBinding(argType) {
							dtoInfo = &DTOInfo{
								Type:        argType,
								TypeName:    ha.extractTypeName(argType),
								Package:     ha.extractPackage(argType),
								ContentType: "application/json",
								IsForm:      false,
							}
						} else if ha.isFormBinding(argType) {
							fields := ha.analyzeFormFields(argType)
							inlineUsages := ha.analyzeFormUsage(funcDecl, info)
							fields = append(fields, inlineUsages...)

							dtoInfo = &DTOInfo{
								Type:        argType,
								TypeName:    ha.extractTypeName(argType),
								Package:     ha.extractPackage(argType),
								ContentType: "multipart/form-data",
								FormFields:  fields,
								IsForm:      true,
							}
						}

						return false // Stop inspection once we find the DTO
					}
				}
			}
		}
		return true
	})

	return dtoInfo
}

// Query params defined on struct
func (ha *HandlerAnalyzer) extractStructQueryParams(t types.Type) []*QueryParam {
	params := make([]*QueryParam, 0)

	// Get underlying struct type
	structType, ok := t.Underlying().(*types.Struct)
	if !ok {
		if ptr, ok := t.(*types.Pointer); ok {
			if st, ok := ptr.Elem().Underlying().(*types.Struct); ok {
				structType = st
			}
		} else {
			return params
		}
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tag := structType.Tag(i)

		isPtr := isPointer(field.Type())
		ftype := field.Type().String()
		if isPtr {
			ftype = ftype[1:] // remove "*"
		}

		// Check for query tag
		if queryTag := reflect.StructTag(tag).Get("query"); queryTag != "" {
			params = append(params, &QueryParam{
				Name:     queryTag,
				Required: !types.IsInterface(field.Type()) && !isPtr,
				Type:     ftype,
			})
		}
	}

	return params
}

func (ha *HandlerAnalyzer) extractTypeName(t types.Type) string {
	if named, ok := t.(*types.Named); ok {
		obj := named.Obj()
		return obj.Pkg().Name() + "." + obj.Name() // pkg.Type
	}

	if ptr, ok := t.(*types.Pointer); ok {
		if named, ok := ptr.Elem().(*types.Named); ok {
			obj := named.Obj()
			return obj.Pkg().Name() + "." + obj.Name() // pkg.Type
		}
	}

	return ""
}

func (ha *HandlerAnalyzer) extractPackage(t types.Type) string {
	if named, ok := t.(*types.Named); ok {
		return named.Obj().Pkg().Path()
	}

	if ptr, ok := t.(*types.Pointer); ok {
		if named, ok := ptr.Elem().(*types.Named); ok {
			return named.Obj().Pkg().Path()
		}
	}

	return ""
}

func (ha *HandlerAnalyzer) analyzeQueryParams(funcDecl *ast.FuncDecl, info *types.Info) []*QueryParam {
	params := make([]*QueryParam, 0)

	// Analyze direct QueryParam calls
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "QueryParam" {
					if len(call.Args) > 0 {
						// Extract param name from first argument
						if lit, ok := call.Args[0].(*ast.BasicLit); ok {
							name := strings.Trim(lit.Value, "\"")
							params = append(params, &QueryParam{
								Name:     name,
								Required: false,
								Type:     paramTypeFromContext(funcDecl, name),
							})
						}
					}
				}
			}
		}
		return true
	})

	// Analyze struct bindings for query params
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Bind" {
					if len(call.Args) > 0 {
						argType := info.TypeOf(call.Args[0])
						if ha.hasQueryTags(argType) {
							params = append(params, ha.extractStructQueryParams(argType)...)
						}
					}
				}
			}
		}
		return true
	})

	return params
}

func (ha *HandlerAnalyzer) isJSONBinding(t types.Type) bool {
	structType, ok := t.Underlying().(*types.Struct)
	if !ok { // TODO: need to write a wrapper
		if ptr, ok := t.(*types.Pointer); ok {
			if st, ok := ptr.Elem().Underlying().(*types.Struct); ok {
				structType = st
			}
		} else {
			return false
		}
	}

	for i := 0; i < structType.NumFields(); i++ {
		tag := structType.Tag(i)
		if jsonTag := reflect.StructTag(tag).Get("json"); jsonTag != "" {
			return true
		}
	}
	return false
}

func (ha *HandlerAnalyzer) hasQueryTags(t types.Type) bool {
	structType, ok := t.Underlying().(*types.Struct)
	if !ok { // TODO: need to write a wrapper
		if ptr, ok := t.(*types.Pointer); ok {
			if st, ok := ptr.Elem().Underlying().(*types.Struct); ok {
				structType = st
			}
		} else {
			return false
		}
	}

	for i := 0; i < structType.NumFields(); i++ {
		tag := structType.Tag(i)
		if queryTag := reflect.StructTag(tag).Get("query"); queryTag != "" {
			return true
		}
	}
	return false
}

// analyzeFormUsage looks for dynamic form usages, like c.FormFile and c.FormValue
func (ha *HandlerAnalyzer) analyzeFormUsage(funcDecl *ast.FuncDecl, info *types.Info) []*FormField {
	fields := make([]*FormField, 0)

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				switch sel.Sel.Name {
				case "FormFile":
					if len(call.Args) > 0 {
						if lit, ok := call.Args[0].(*ast.BasicLit); ok {
							fields = append(fields, &FormField{
								Name:     strings.Trim(lit.Value, "\""),
								Type:     "string",
								Required: false,
								IsFile:   true,
								IsArray:  false,
							})
						}
					}

				case "FormValue":
					if len(call.Args) > 0 {
						if lit, ok := call.Args[0].(*ast.BasicLit); ok {
							fields = append(fields, &FormField{
								Name:     strings.Trim(lit.Value, "\""),
								Type:     "string", // TODO: extract exact type
								Required: false,
								IsFile:   false,
								IsArray:  false,
							})
						}
					}
				}
			}
		}
		return true
	})

	return fields
}

func (ha *HandlerAnalyzer) isFormBinding(t types.Type) bool {
	structType, ok := t.Underlying().(*types.Struct)
	if !ok { // TODO: need to write a wrapper
		if ptr, ok := t.(*types.Pointer); ok {
			if st, ok := ptr.Elem().Underlying().(*types.Struct); ok {
				structType = st
			}
		} else {
			return false
		}
	}

	for i := 0; i < structType.NumFields(); i++ {
		tag := structType.Tag(i)
		if formTag := reflect.StructTag(tag).Get("form"); formTag != "" {
			return true
		}
	}
	return false
}

// analyzeFormFields extracts form fields from struct with "form" tag
func (ha *HandlerAnalyzer) analyzeFormFields(t types.Type) []*FormField {
	fields := make([]*FormField, 0)

	structType, ok := t.Underlying().(*types.Struct)
	if !ok { // TODO: need to write a wrapper
		if ptr, ok := t.(*types.Pointer); ok {
			if st, ok := ptr.Elem().Underlying().(*types.Struct); ok {
				structType = st
			}
		} else {
			return fields
		}
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tag := structType.Tag(i)

		if formTag := reflect.StructTag(tag).Get("form"); formTag != "" {
			isFile, isArray := isFileField(field.Type())
			fieldType := "string" // TODO: extract exact type, separate format!!!
			if isFile {
				fieldType = "string"
			}

			fields = append(fields, &FormField{
				Name:     formTag,
				Type:     fieldType,
				Required: !isPointer(field.Type()),
				IsFile:   isFile,
				IsArray:  isArray,
			})
		}
	}

	return fields
}

func isFileField(t types.Type) (isFile bool, isArray bool) {
	const mimePkgPath = "mime/multipart"

	// Check for single file
	if ptr, ok := t.(*types.Pointer); ok {
		if named, ok := ptr.Elem().(*types.Named); ok {
			return named.Obj().Pkg().Path() == mimePkgPath && named.Obj().Name() == "FileHeader", false
		}
	}

	// Check for file array
	if slice, ok := t.(*types.Slice); ok {
		if ptr, ok := slice.Elem().(*types.Pointer); ok {
			if named, ok := ptr.Elem().(*types.Named); ok {
				return named.Obj().Pkg().Path() == mimePkgPath && named.Obj().Name() == "FileHeader", true
			}
		}
	}

	return false, false
}

func (ha *HandlerAnalyzer) extractDocumentation(funcDecl *ast.FuncDecl) string {
	if funcDecl.Doc == nil {
		return ""
	}

	var doc strings.Builder
	for _, comment := range funcDecl.Doc.List {
		doc.WriteString(strings.TrimPrefix(comment.Text, "// "))
		doc.WriteString("\n")
	}

	return strings.TrimSpace(doc.String())
}

func isPointer(t types.Type) bool {
	_, ok := t.(*types.Pointer)
	return ok
}

func (ha *HandlerAnalyzer) buildPatterns() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current working directory: %v", err)
	}

	var patterns []string
	for _, pkg := range ha.pkgs {
		pattern := pkg.Path
		if pkg.Recursive {
			pattern = filepath.Join(cwd, pattern, "...")
		}

		patterns = append(patterns, pattern)
	}

	return patterns, nil
}
