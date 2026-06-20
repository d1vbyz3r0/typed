package request

import (
	"mime/multipart"
	"reflect"
	"testing"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/testsuite"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func requestFromFixture(t *testing.T, fixture string) *Request {
	t.Helper()

	pkg, fn := testsuite.LoadFixtureFunc(t, "request/"+fixture, "Handler")
	return New(fn, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
}

func TestNewRequest_JSON(t *testing.T) {
	want := &Request{
		ModelType: typing.Named("github.com/d1vbyz3r0/typed/testdata/request/jsontest", "JsonDTO"),
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationJSON: Body{},
		},
	}

	require.Equal(t, want, requestFromFixture(t, "jsontest"))
}

func TestNewRequest_XML(t *testing.T) {
	want := &Request{
		ModelType: typing.Named("github.com/d1vbyz3r0/typed/testdata/request/xmltest", "XMLDto"),
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationXML: Body{},
		},
	}

	require.Equal(t, want, requestFromFixture(t, "xmltest"))
}

func TestNewRequest_EmptyTags(t *testing.T) {
	want := &Request{
		ModelType: typing.Named("github.com/d1vbyz3r0/typed/testdata/request/emptytest", "NoTags"),
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationJSON: Body{},
			echo.MIMEApplicationXML:  Body{},
		},
	}

	require.Equal(t, want, requestFromFixture(t, "emptytest"))
}

func TestNewRequest_FormTagsNoFiles(t *testing.T) {
	want := &Request{
		ModelType: typing.Named("github.com/d1vbyz3r0/typed/testdata/request/formtest/nofile", "Form"),
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationForm: Body{},
			echo.MIMEMultipartForm:   Body{},
		},
	}

	require.Equal(t, want, requestFromFixture(t, "formtest/nofile"))
}

func TestNewRequest_FormWithFile(t *testing.T) {
	want := &Request{
		ModelType: typing.Named("github.com/d1vbyz3r0/typed/testdata/request/formtest/file", "Form"),
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEMultipartForm: Body{},
		},
	}

	require.Equal(t, want, requestFromFixture(t, "formtest/file"))
}

func TestNewRequest_FormWithFiles(t *testing.T) {
	want := &Request{
		ModelType: typing.Named("github.com/d1vbyz3r0/typed/testdata/request/formtest/files", "Form"),
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEMultipartForm: Body{},
		},
	}

	require.Equal(t, want, requestFromFixture(t, "formtest/files"))
}

func TestNewRequest_NoBody(t *testing.T) {
	want := &Request{
		ContentTypeMapping: ContentTypeMapping{},
	}

	require.Equal(t, want, requestFromFixture(t, "nobody"))
}

func TestNewRequest_NoBinds(t *testing.T) {
	want := &Request{
		ContentTypeMapping: ContentTypeMapping{},
	}

	require.Equal(t, want, requestFromFixture(t, "nobind"))
}

func TestNewRequest_MultipleTags(t *testing.T) {
	want := &Request{
		ModelType: typing.Named("github.com/d1vbyz3r0/typed/testdata/request/multiple", "Data"),
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationJSON: Body{},
			echo.MIMEMultipartForm:   Body{},
			echo.MIMEApplicationForm: Body{},
			echo.MIMEApplicationXML:  Body{},
		},
	}

	require.Equal(t, want, requestFromFixture(t, "multiple"))
}

func TestNewRequest_InlineFormNoFiles(t *testing.T) {
	f := reflect.StructOf([]reflect.StructField{
		{
			Name: "Name",
			Type: reflect.TypeFor[string](),
			Tag:  `form:"name"`,
		},
		{
			Name: "Age",
			Type: reflect.TypeFor[int](),
			Tag:  `form:"age"`,
		},
	})

	want := &Request{
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationForm: Body{
				Form: f,
			},
			echo.MIMEMultipartForm: Body{
				Form: f,
			},
		},
	}

	require.Equal(t, want, requestFromFixture(t, "formtest/inline"))
}

func TestNewRequest_InlineFormWithFiles(t *testing.T) {
	f := reflect.StructOf([]reflect.StructField{
		{
			Name: "Name",
			Type: reflect.TypeFor[string](),
			Tag:  `form:"name"`,
		},
		{
			Name: "File",

			Type: reflect.TypeFor[*multipart.FileHeader](),
			Tag:  `form:"file"`,
		},
	})

	want := &Request{
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEMultipartForm: Body{
				Form: f,
			},
		},
	}

	req := requestFromFixture(t, "formtest/inlinefiles")
	require.Equal(t, want.ModelType, req.ModelType)
	got := req.ContentTypeMapping[echo.MIMEMultipartForm].Form
	require.Equal(t, f, got)
}
