package format

import (
	"github.com/d1vbyz3r0/typed/common/typing"
	"log/slog"
	"reflect"
)

// https://github.com/go-playground/validator/blob/master/regexes.go

const (
	AlphaRegexString                 = "[a-zA-Z]+"
	AlphaNumericRegexString          = "[a-zA-Z0-9]+"
	AlphaUnicodeRegexString          = "[\\p{L}]+"
	AlphaUnicodeNumericRegexString   = "[\\p{L}\\p{N}]+"
	NumericRegexString               = "[-+]?[0-9]+(?:\\.[0-9]+)?"
	NumberRegexString                = "[0-9]+"
	HexadecimalRegexString           = "(0[xX])?[0-9a-fA-F]+"
	HexColorRegexString              = "#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})"
	RgbRegexString                   = "rgb\\(\\s*(?:(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])|(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%)\\s*\\)"
	RgbaRegexString                  = "rgba\\(\\s*(?:(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])|(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%)\\s*,\\s*(?:(?:0.[1-9]*)|[01])\\s*\\)"
	HslRegexString                   = "hsl\\(\\s*(?:0|[1-9]\\d?|[12]\\d\\d|3[0-5]\\d|360)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*\\)"
	HslaRegexString                  = "hsla\\(\\s*(?:0|[1-9]\\d?|[12]\\d\\d|3[0-5]\\d|360)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*,\\s*(?:(?:0.[1-9]*)|[01])\\s*\\)"
	EmailRegexString                 = "(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?"
	E164RegexString                  = "\\+[1-9]?[0-9]{7,14}"
	Base32RegexString                = "(?:[A-Z2-7]{8})*(?:[A-Z2-7]{2}={6}|[A-Z2-7]{4}={4}|[A-Z2-7]{5}={3}|[A-Z2-7]{7}=|[A-Z2-7]{8})"
	Base64RegexString                = "(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})"
	Base64URLRegexString             = "(?:[A-Za-z0-9-_]{4})*(?:[A-Za-z0-9-_]{2}==|[A-Za-z0-9-_]{3}=|[A-Za-z0-9-_]{4})"
	Base64RawURLRegexString          = "(?:[A-Za-z0-9-_]{4})*(?:[A-Za-z0-9-_]{2,4})"
	ISBN10RegexString                = "(?:[0-9]{9}X|[0-9]{10})"
	ISBN13RegexString                = "(?:(?:97(?:8|9))[0-9]{10})"
	ISSNRegexString                  = "(?:[0-9]{4}-[0-9]{3}[0-9X])"
	UUID3RegexString                 = "[0-9a-f]{8}-[0-9a-f]{4}-3[0-9a-f]{3}-[0-9a-f]{4}-[0-9a-f]{12}"
	UUID4RegexString                 = "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}"
	UUID5RegexString                 = "[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}"
	UUIDRegexString                  = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
	UUID3RFC4122RegexString          = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-3[0-9a-fA-F]{3}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
	UUID4RFC4122RegexString          = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}"
	UUID5RFC4122RegexString          = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-5[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}"
	UUIDRFC4122RegexString           = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
	ULIDRegexString                  = "(?i)[A-HJKMNP-TV-Z0-9]{26}"
	MD4RegexString                   = "[0-9a-f]{32}"
	MD5RegexString                   = "[0-9a-f]{32}"
	SHA256RegexString                = "[0-9a-f]{64}"
	SHA384RegexString                = "[0-9a-f]{96}"
	SHA512RegexString                = "[0-9a-f]{128}"
	Ripemd128RegexString             = "[0-9a-f]{32}"
	Ripemd160RegexString             = "[0-9a-f]{40}"
	Tiger128RegexString              = "[0-9a-f]{32}"
	Tiger160RegexString              = "[0-9a-f]{40}"
	Tiger192RegexString              = "[0-9a-f]{48}"
	ASCIIRegexString                 = "[\x00-\x7F]*"
	PrintableASCIIRegexString        = "[\x20-\x7E]*"
	MultibyteRegexString             = "[^\x00-\x7F]"
	DataURIRegexString               = `data:((?:\w+\/(?:([^;]|;[^;]).)+)?)`
	LatitudeRegexString              = "[-+]?([1-8]?\\d(\\.\\d+)?|90(\\.0+)?)"
	LongitudeRegexString             = "[-+]?(180(\\.0+)?|((1[0-7]\\d)|([1-9]?\\d))(\\.\\d+)?)"
	SSNRegexString                   = `[0-9]{3}[ -]?(0[1-9]|[1-9][0-9])[ -]?([1-9][0-9]{3}|[0-9][1-9][0-9]{2}|[0-9]{2}[1-9][0-9]|[0-9]{3}[1-9])`
	HostnameRegexStringRFC952        = `[a-zA-Z]([a-zA-Z0-9\-]+[\.]?)*[a-zA-Z0-9]`                                                                   // https://tools.ietf.org/html/rfc952
	HostnameRegexStringRFC1123       = `([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62}){1}(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?`                                 // accepts hostname starting with a digit https://tools.ietf.org/html/rfc1123
	FqdnRegexStringRFC1123           = `([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?(\.[a-zA-Z]{1}[a-zA-Z0-9]{0,62})\.?` // same as hostnameRegexStringRFC1123 but must contain a non numerical TLD (possibly ending with '.')
	BtcAddressRegexString            = `[13][a-km-zA-HJ-NP-Z1-9]{25,34}`                                                                             // bitcoin address
	BtcAddressUpperRegexStringBech32 = `BC1[02-9AC-HJ-NP-Z]{7,76}`                                                                                   // bitcoin bech32 address https://en.bitcoin.it/wiki/Bech32
	BtcAddressLowerRegexStringBech32 = `bc1[02-9ac-hj-np-z]{7,76}`                                                                                   // bitcoin bech32 address https://en.bitcoin.it/wiki/Bech32
	EthAddressRegexString            = `0x[0-9a-fA-F]{40}`
	EthAddressUpperRegexString       = `0x[0-9A-F]{40}`
	EthAddressLowerRegexString       = `0x[0-9a-f]{40}`
	URLEncodedRegexString            = `(?:[^%]|%[0-9A-Fa-f]{2})*`
	HTMLEncodedRegexString           = `&#[x]?([0-9a-fA-F]{2})|(&gt)|(&lt)|(&quot)|(&amp)+[;]?`
	HTMLRegexString                  = `<[/]?([a-zA-Z]+).*?>`
	JWTRegexString                   = "[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]*"
	BicRegexString                   = `[A-Za-z]{6}[A-Za-z0-9]{2}([A-Za-z0-9]{3})?`
	SemverRegexString                = `(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?` // numbered capture groups https://semver.org/
	DNSRegexStringRFC1035Label       = "[a-z]([-a-z0-9]*[a-z0-9])?"
	CVERegexString                   = `CVE-(1999|2\d{3})-(0[^0]\d{2}|0\d[^0]\d{1}|0\d{2}[^0]|[1-9]{1}\d{3,})` // CVE Format Id https://cve.mitre.org/cve/identifiers/syntaxchange.html
	MongodbIdRegexString             = "[a-f\\d]{24}"
	MongodbConnStringRegexString     = "mongodb(\\+srv)?:\\/\\/(([a-zA-Z\\d]+):([a-zA-Z\\d$:\\/?#\\[\\]@]+)@)?(([a-z\\d.-]+)(:[\\d]+)?)((,(([a-z\\d.-]+)(:(\\d+))?))*)?(\\/[a-zA-Z-_]{1,64})?(\\?(([a-zA-Z]+)=([a-zA-Z\\d]+))(&(([a-zA-Z\\d]+)=([a-zA-Z\\d]+))?)*)?"
	CronRegexString                  = `(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|((\*|\d+)(\/|-)\d+)|\d+|\*) ?){5,7})`
	SpicedbIDRegexString             = `(([a-zA-Z0-9/_|\-=+]{1,})|\*)`
	SpicedbPermissionRegexString     = "([a-z][a-z0-9_]{1,62}[a-z0-9])?"
	SpicedbTypeRegexString           = "([a-z][a-z0-9_]{1,61}[a-z0-9]/)?[a-z][a-z0-9_]{1,62}[a-z0-9]"
	EinRegexString                   = "(\\d{2}-\\d{7})"
)

// additional regexes not shipped with https://github.com/go-playground/validator
const (
	CIDRRegex                   = "([0-9]{1,3}\\.){3}[0-9]{1,3}\\/[0-9]{1,2}"
	MACRegexString              = "([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})"
	TCP4AddrRegexString         = "([0-9]{1,3}\\.){3}[0-9]{1,3}:[0-9]+"
	TCP6AddrRegexString         = "\\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\\]:[0-9]+"
	TCPAddrRegexString          = "(([0-9]{1,3}\\.){3}[0-9]{1,3}|\\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\\]):[0-9]+"
	UDP4AddrRegexString         = "([0-9]{1,3}\\.){3}[0-9]{1,3}:[0-9]+"
	UDP6AddrRegexString         = "\\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\\]:[0-9]+"
	UDPAddrRegexString          = "(([0-9]{1,3}\\.){3}[0-9]{1,3}|\\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\\]):[0-9]+"
	UnixAddrRegexString         = "/.+/.+"
	URNRFC2141RegexString       = "urn:[a-zA-Z0-9][a-zA-Z0-9-]{0,31}:([a-zA-Z0-9()+,\\-.:=@;$_!*'%/?#]|%[0-9a-fA-F]{2})+"
	ISBNRegexString             = `(?:[0-9]{9}[\dXx]|[0-9]{13})`
	BtcAddressBech32RegexString = `(?i)(bc1)[0-9ac-hj-np-z]{25,39}`
	IPAnyRegexString            = `((([0-9]{1,3}\.){3}[0-9]{1,3})|(([a-fA-F0-9:]+)))`
	NonEmptyRegexString         = ".+"
)

var Formats = map[string]TagFn{
	"alpha": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(AlphaRegexString)
	},
	"alphanum": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(AlphaNumericRegexString)
	},
	"alphaunicode": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(AlphaUnicodeRegexString)
	},
	"alphanumunicode": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(AlphaUnicodeNumericRegexString)
	},
	"numeric": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(NumericRegexString)
	},
	"number": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(NumberRegexString)
	},
	"hexadeсimal": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(HexadecimalRegexString)
	},
	"hexcolor": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(HexColorRegexString)
	},
	"rgb": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(RgbRegexString)
	},
	"rgba": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(RgbaRegexString)
	},
	"hsl": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(HslRegexString)
	},
	"hsla": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(HslaRegexString)
	},
	"email": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Format = Email
	},
	"e164": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(E164RegexString)
	},
	"base32": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(Base32RegexString)
	},
	"base64": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(Base64RegexString)
	},
	"base64url": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(Base64URLRegexString)
	},
	"base64rawurl": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(Base64RawURLRegexString)
	},
	"isbn": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(ISBNRegexString)
	},
	"isbn10": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(ISBN10RegexString)
	},
	"isbn13": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(ISBN13RegexString)
	},
	"issn": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(ISSNRegexString)
	},
	"uuid": func(ctx *FieldContext) {
		ctx.Required = true
		// TODO: what if has or ???
		ctx.Format = UUID
	},
	"uuid3": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(UUID3RegexString)
	},
	"uuid4": func(ctx *FieldContext) {
		ctx.Required = true
		// TODO: what if has or ???
		ctx.Format = UUID
	},
	"uuid5": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(UUID5RegexString)
	},
	"uuid_rfc4122": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(UUIDRFC4122RegexString)
	},
	"uuid3_rfc4122": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(UUID3RFC4122RegexString)
	},
	"uuid4_rfc4122": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(UUID4RFC4122RegexString)
	},
	"uuid5_rfc4122": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(UUID5RFC4122RegexString)
	},
	"ulid": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(ULIDRegexString)
	},
	"md4": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(MD4RegexString)
	},
	"md5": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(MD5RegexString)
	},
	"sha256": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(SHA256RegexString)
	},
	"sha384": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(SHA384RegexString)
	},
	"sha512": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(SHA512RegexString)
	},
	"ripemd128": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(Ripemd128RegexString)
	},
	"ripemd160": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(Ripemd160RegexString)
	},
	"tiger128": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(Tiger128RegexString)
	},
	"tiger160": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(Tiger160RegexString)
	},
	"tiger192": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Pattern = Tiger192RegexString
	},
	"ascii": func(ctx *FieldContext) {
		ctx.Pattern = ASCIIRegexString
	},
	"printascii": func(ctx *FieldContext) {
		ctx.Pattern = PrintableASCIIRegexString
	},
	"multibyte": func(ctx *FieldContext) {
		ctx.Pattern = MultibyteRegexString
	},
	"datauri": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Pattern = DataURIRegexString
	},
	"latitude": func(ctx *FieldContext) {
		ctx.Required = true
		if ctx.Type.Kind() == reflect.String {
			ctx.AddPattern(LatitudeRegexString)
		} else {
			_min := float64(-90)
			_max := float64(90)
			ctx.Min = &_min
			ctx.Min = &_max
		}
	},
	"longitude": func(ctx *FieldContext) {
		ctx.Required = true
		if ctx.Type.Kind() == reflect.String {
			ctx.AddPattern(LongitudeRegexString)
		} else {
			_min := float64(-180)
			_max := float64(180)
			ctx.Min = &_min
			ctx.Min = &_max
		}
	},
	"ssn": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Pattern = SSNRegexString
	},
	"hostname": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(HostnameRegexStringRFC952)
	},
	"hostname_rfc1123": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Format = Hostname
	},
	"fqdn": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(FqdnRegexStringRFC1123)
	},
	"btc_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(BtcAddressRegexString)
	},
	"btc_addr_bech32": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(BtcAddressBech32RegexString)
	},
	"eth_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(EthAddressRegexString)
	},
	"url_encoded": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(URLEncodedRegexString)
	},
	"html": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(HTMLRegexString)
	},
	"html_encoded": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(HTMLEncodedRegexString)
	},
	"jwt": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(JWTRegexString)
	},
	"bic": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(BicRegexString)
	},
	"semver": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(SemverRegexString)
	},
	"dns_rfc1035_label": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(DNSRegexStringRFC1035Label)
	},
	"cve": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(CVERegexString)
	},
	"mongodb": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(MongodbIdRegexString)
	},
	"mongodb_connection_string": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(MongodbConnStringRegexString)
	},
	"cron": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(CronRegexString)
	},
	"spicedb": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(SpicedbIDRegexString)
	},
	"ein": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(EinRegexString)
	},

	"cidr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(CIDRRegex)
	},
	"cidrv4": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(CIDRRegex)
	},
	"hostname_port": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Format = Hostname
	},
	"ip": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(IPAnyRegexString)
	},
	"ip_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(IPAnyRegexString)
	},
	"ipv4": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Format = IPv4
	},
	"ip4_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Format = IPv4
	},
	"ipv6": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Format = IPv6
	},
	"ip6_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Format = IPv6
	},
	"mac": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(MACRegexString)
	},
	"tcp4_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(TCP4AddrRegexString)
	},
	"tcp6_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(TCP6AddrRegexString)
	},
	"tcp_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(TCPAddrRegexString)
	},
	"udp4_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(UDP4AddrRegexString)
	},
	"udp6_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(UDP6AddrRegexString)
	},
	"udp_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(UDPAddrRegexString)
	},
	"unix_addr": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(UnixAddrRegexString)
	},
	"uri": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Format = URI
	},
	"url": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.Format = URI
	},
	"urn_rfc2141": func(ctx *FieldContext) {
		ctx.Required = true
		ctx.AddPattern(URNRFC2141RegexString)
	},

	"required": func(ctx *FieldContext) {
		ctx.Required = true
	},

	"lt": func(ctx *FieldContext) {
		lt, err := ctx.LookupFloat("lt")
		if err != nil {
			return
		}
		ctx.Required = true
		ctx.Max = &lt
		ctx.ExclusiveMax = true
	},
	"lte": func(ctx *FieldContext) {
		lte, err := ctx.LookupFloat("lte")
		if err != nil {
			return
		}
		ctx.Required = true
		ctx.Max = &lte
		ctx.ExclusiveMax = false
	},
	"gt": func(ctx *FieldContext) {
		ctx.Required = true
		gt, err := ctx.LookupFloat("gt")
		if err != nil {
			return
		}
		ctx.Required = true
		ctx.Min = &gt
		ctx.ExclusiveMin = true
	},
	"gte": func(ctx *FieldContext) {
		ctx.Required = true
		gte, err := ctx.LookupFloat("gte")
		if err != nil {
			return
		}
		ctx.Required = true
		ctx.Min = &gte
		ctx.ExclusiveMin = false
	},

	"eq": func(ctx *FieldContext) {
		floatType := reflect.TypeOf(float64(0))
		stringType := reflect.TypeOf("")
		if ctx.Type.ConvertibleTo(floatType) {

		} else if ctx.Type.ConvertibleTo(stringType) {

		}

		ctx.Required = true
	},

	"oneof": func(ctx *FieldContext) {
		ctx.Required = true
		floatType := reflect.TypeOf(float64(0))
		stringType := reflect.TypeOf("")

		if ctx.Type.ConvertibleTo(floatType) {
			v, err := ctx.LookupFloatSlice("oneof")
			if err != nil {
				slog.Warn("lookup float oneof value", "error", err)
				return
			}
			ctx.OneOf = typing.EraseSliceType(v)
		} else if ctx.Type.ConvertibleTo(stringType) {
			v, err := ctx.LookupStringSlice("oneof")
			if err != nil {
				slog.Warn("lookup string oneof value", "error", err)
				return
			}
			ctx.OneOf = typing.EraseSliceType(v)
		}
	},
}
