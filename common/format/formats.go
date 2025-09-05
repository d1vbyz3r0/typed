package format

// https://github.com/go-playground/validator/blob/master/regexes.go

const (
	AlphaRegexString               = "^[a-zA-Z]+$"
	AlphaNumericRegexString        = "^[a-zA-Z0-9]+$"
	AlphaUnicodeRegexString        = "^[\\p{L}]+$"
	AlphaUnicodeNumericRegexString = "^[\\p{L}\\p{N}]+$"
	NumericRegexString             = "^[-+]?[0-9]+(?:\\.[0-9]+)?$"
	NumberRegexString              = "^[0-9]+$"
	HexadecimalRegexString         = "^(0[xX])?[0-9a-fA-F]+$"
	HexColorRegexString            = "^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$"
	RgbRegexString                 = "^rgb\\(\\s*(?:(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])|(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%)\\s*\\)$"
	RgbaRegexString                = "^rgba\\(\\s*(?:(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])|(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%)\\s*,\\s*(?:(?:0.[1-9]*)|[01])\\s*\\)$"
	HslRegexString                 = "^hsl\\(\\s*(?:0|[1-9]\\d?|[12]\\d\\d|3[0-5]\\d|360)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*\\)$"
	HslaRegexString                = "^hsla\\(\\s*(?:0|[1-9]\\d?|[12]\\d\\d|3[0-5]\\d|360)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*,\\s*(?:(?:0.[1-9]*)|[01])\\s*\\)$"
	// emailRegexString                 = "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
	E164RegexString            = "^\\+[1-9]?[0-9]{7,14}$"
	Base32RegexString          = "^(?:[A-Z2-7]{8})*(?:[A-Z2-7]{2}={6}|[A-Z2-7]{4}={4}|[A-Z2-7]{5}={3}|[A-Z2-7]{7}=|[A-Z2-7]{8})$"
	Base64RegexString          = "^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$"
	Base64URLRegexString       = "^(?:[A-Za-z0-9-_]{4})*(?:[A-Za-z0-9-_]{2}==|[A-Za-z0-9-_]{3}=|[A-Za-z0-9-_]{4})$"
	Base64RawURLRegexString    = "^(?:[A-Za-z0-9-_]{4})*(?:[A-Za-z0-9-_]{2,4})$"
	ISBN10RegexString          = "^(?:[0-9]{9}X|[0-9]{10})$"
	ISBN13RegexString          = "^(?:(?:97(?:8|9))[0-9]{10})$"
	ISSNRegexString            = "^(?:[0-9]{4}-[0-9]{3}[0-9X])$"
	UUID3RegexString           = "^[0-9a-f]{8}-[0-9a-f]{4}-3[0-9a-f]{3}-[0-9a-f]{4}-[0-9a-f]{12}$"
	UUID4RegexString           = "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
	UUID5RegexString           = "^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
	UUIDRegexString            = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	UUID3RFC4122RegexString    = "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-3[0-9a-fA-F]{3}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"
	UUID4RFC4122RegexString    = "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$"
	UUID5RFC4122RegexString    = "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-5[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$"
	UUIDRFC4122RegexString     = "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"
	ULIDRegexString            = "^(?i)[A-HJKMNP-TV-Z0-9]{26}$"
	MD4RegexString             = "^[0-9a-f]{32}$"
	MD5RegexString             = "^[0-9a-f]{32}$"
	SHA256RegexString          = "^[0-9a-f]{64}$"
	SHA384RegexString          = "^[0-9a-f]{96}$"
	SHA512RegexString          = "^[0-9a-f]{128}$"
	Ripemd128RegexString       = "^[0-9a-f]{32}$"
	Ripemd160RegexString       = "^[0-9a-f]{40}$"
	Tiger128RegexString        = "^[0-9a-f]{32}$"
	Tiger160RegexString        = "^[0-9a-f]{40}$"
	Tiger192RegexString        = "^[0-9a-f]{48}$"
	ASCIIRegexString           = "^[\x00-\x7F]*$"
	PrintableASCIIRegexString  = "^[\x20-\x7E]*$"
	MultibyteRegexString       = "[^\x00-\x7F]"
	DataURIRegexString         = `^data:((?:\w+\/(?:([^;]|;[^;]).)+)?)`
	LatitudeRegexString        = "^[-+]?([1-8]?\\d(\\.\\d+)?|90(\\.0+)?)$"
	LongitudeRegexString       = "^[-+]?(180(\\.0+)?|((1[0-7]\\d)|([1-9]?\\d))(\\.\\d+)?)$"
	SSNRegexString             = `^[0-9]{3}[ -]?(0[1-9]|[1-9][0-9])[ -]?([1-9][0-9]{3}|[0-9][1-9][0-9]{2}|[0-9]{2}[1-9][0-9]|[0-9]{3}[1-9])$`
	HostnameRegexStringRFC952  = `^[a-zA-Z]([a-zA-Z0-9\-]+[\.]?)*[a-zA-Z0-9]$`                                                                   // https://tools.ietf.org/html/rfc952
	HostnameRegexStringRFC1123 = `^([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62}){1}(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?$`                                 // accepts hostname starting with a digit https://tools.ietf.org/html/rfc1123
	FqdnRegexStringRFC1123     = `^([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?(\.[a-zA-Z]{1}[a-zA-Z0-9]{0,62})\.?$` // same as hostnameRegexStringRFC1123 but must contain a non numerical TLD (possibly ending with '.')
	BtcAddressRegexString      = `^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`                                                                             // bitcoin address
	// btcAddressUpperRegexStringBech32 = `^BC1[02-9AC-HJ-NP-Z]{7,76}$`                                                                                   // bitcoin bech32 address https://en.bitcoin.it/wiki/Bech32
	// btcAddressLowerRegexStringBech32 = `^bc1[02-9ac-hj-np-z]{7,76}$`                                                                                   // bitcoin bech32 address https://en.bitcoin.it/wiki/Bech32
	EthAddressRegexString = `^0x[0-9a-fA-F]{40}$`
	// ethAddressUpperRegexString       = `^0x[0-9A-F]{40}$`
	// ethAddressLowerRegexString       = `^0x[0-9a-f]{40}$`
	URLEncodedRegexString  = `^(?:[^%]|%[0-9A-Fa-f]{2})*$`
	HTMLEncodedRegexString = `&#[x]?([0-9a-fA-F]{2})|(&gt)|(&lt)|(&quot)|(&amp)+[;]?`
	HTMLRegexString        = `<[/]?([a-zA-Z]+).*?>`
	JWTRegexString         = "^[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]*$"
	// splitParamsRegexString           = `'[^']*'|\S+`
	BicRegexString               = `^[A-Za-z]{6}[A-Za-z0-9]{2}([A-Za-z0-9]{3})?$`
	SemverRegexString            = `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$` // numbered capture groups https://semver.org/
	DNSRegexStringRFC1035Label   = "^[a-z]([-a-z0-9]*[a-z0-9])?$"
	CVERegexString               = `^CVE-(1999|2\d{3})-(0[^0]\d{2}|0\d[^0]\d{1}|0\d{2}[^0]|[1-9]{1}\d{3,})$` // CVE Format Id https://cve.mitre.org/cve/identifiers/syntaxchange.html
	MongodbIdRegexString         = "^[a-f\\d]{24}$"
	MongodbConnStringRegexString = "^mongodb(\\+srv)?:\\/\\/(([a-zA-Z\\d]+):([a-zA-Z\\d$:\\/?#\\[\\]@]+)@)?(([a-z\\d.-]+)(:[\\d]+)?)((,(([a-z\\d.-]+)(:(\\d+))?))*)?(\\/[a-zA-Z-_]{1,64})?(\\?(([a-zA-Z]+)=([a-zA-Z\\d]+))(&(([a-zA-Z\\d]+)=([a-zA-Z\\d]+))?)*)?$"
	CronRegexString              = `(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|((\*|\d+)(\/|-)\d+)|\d+|\*) ?){5,7})`
	SpicedbIDRegexString         = `^(([a-zA-Z0-9/_|\-=+]{1,})|\*)$`
	// spicedbPermissionRegexString = "^([a-z][a-z0-9_]{1,62}[a-z0-9])?$"
	// spicedbTypeRegexString       = "^([a-z][a-z0-9_]{1,61}[a-z0-9]/)?[a-z][a-z0-9_]{1,62}[a-z0-9]$"
	EinRegexString = "^(\\d{2}-\\d{7})$"
)

// additional regexes not shipped with https://github.com/go-playground/validator
const (
	CIDRRegex                   = "^([0-9]{1,3}\\.){3}[0-9]{1,3}\\/[0-9]{1,2}$"
	MACRegexString              = "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"
	TCP4AddrRegexString         = "^([0-9]{1,3}\\.){3}[0-9]{1,3}:[0-9]+$"
	TCP6AddrRegexString         = "^\\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\\]:[0-9]+$"
	TCPAddrRegexString          = "^(([0-9]{1,3}\\.){3}[0-9]{1,3}|\\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\\]):[0-9]+$"
	UDP4AddrRegexString         = "^([0-9]{1,3}\\.){3}[0-9]{1,3}:[0-9]+$"
	UDP6AddrRegexString         = "^\\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\\]:[0-9]+$"
	UDPAddrRegexString          = "^(([0-9]{1,3}\\.){3}[0-9]{1,3}|\\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\\]):[0-9]+$"
	UnixAddrRegexString         = "^/.+/.+"
	URNRFC2141RegexString       = "^urn:[a-zA-Z0-9][a-zA-Z0-9-]{0,31}:([a-zA-Z0-9()+,\\-.:=@;$_!*'%/?#]|%[0-9a-fA-F]{2})+$"
	ISBNRegexString             = `^(?:[0-9]{9}[\dXx]|[0-9]{13})$`
	BtcAddressBech32RegexString = `^(?i)(bc1)[0-9ac-hj-np-z]{25,39}$`
	IPAnyRegexString            = `^((([0-9]{1,3}\.){3}[0-9]{1,3})|(([a-fA-F0-9:]+)))$`
	NonEmptyRegexString         = "^.+$"
)

// Formats contains binding of validate tag to Meta info
var Formats = map[string]Meta{
	"alpha":           {Pattern: AlphaRegexString},
	"alphanum":        {Pattern: AlphaNumericRegexString},
	"alphaunicode":    {Pattern: AlphaUnicodeRegexString},
	"alphanumunicode": {Pattern: AlphaUnicodeNumericRegexString},
	"numeric":         {Pattern: NumericRegexString},
	"number":          {Pattern: NumberRegexString},
	"hexadeсimal":     {Pattern: HexadecimalRegexString},
	"hexcolor":        {Pattern: HexColorRegexString},
	"rgb":             {Pattern: RgbRegexString},
	"rgba":            {Pattern: RgbaRegexString},
	"hsl":             {Pattern: HslRegexString},
	"hsla":            {Pattern: HslaRegexString},
	"email":           {Format: Email},
	"e164":            {Pattern: E164RegexString},
	"base32":          {Pattern: Base32RegexString},
	"base64":          {Pattern: Base64RegexString},
	"base64url":       {Pattern: Base64URLRegexString},
	"base64rawurl":    {Pattern: Base64RawURLRegexString},
	"isbn":            {Pattern: ISBNRegexString},
	"isbn10":          {Pattern: ISBN10RegexString},
	"isbn13":          {Pattern: ISBN13RegexString},
	"issn":            {Pattern: ISSNRegexString},
	"uuid":            {Format: UUID},
	"uuid3":           {Pattern: UUID3RegexString},
	"uuid4":           {Pattern: UUID4RegexString},
	"uuid5":           {Pattern: UUID5RegexString},
	"uuid_rfc4122":    {Pattern: UUIDRFC4122RegexString},
	"uuid3_rfc4122":   {Pattern: UUID3RFC4122RegexString},
	"uuid4_rfc4122":   {Pattern: UUID4RFC4122RegexString},
	"uuid5_rfc4122":   {Pattern: UUID5RFC4122RegexString},
	"ulid":            {Pattern: ULIDRegexString},
	"md4":             {Pattern: MD4RegexString},
	"md5":             {Pattern: MD5RegexString},
	"sha256":          {Pattern: SHA256RegexString},
	"sha384":          {Pattern: SHA384RegexString},
	"sha512":          {Pattern: SHA512RegexString},
	"ripemd128":       {Pattern: Ripemd128RegexString},
	"ripemd160":       {Pattern: Ripemd160RegexString},
	"tiger128":        {Pattern: Tiger128RegexString},
	"tiger160":        {Pattern: Tiger160RegexString},
	"tiger192":        {Pattern: Tiger192RegexString},
	"ascii":           {Pattern: ASCIIRegexString},
	"printascii":      {Pattern: PrintableASCIIRegexString},
	"multibyte":       {Pattern: MultibyteRegexString},
	"datauri":         {Pattern: DataURIRegexString},
	"latitude":        {Pattern: LatitudeRegexString},
	"longitude":       {Pattern: LongitudeRegexString},
	"ssn":             {Pattern: SSNRegexString},
	"hostname": {
		Pattern: HostnameRegexStringRFC952,
		Format:  Hostname,
	},
	"hostname_rfc1123": {
		Pattern: HostnameRegexStringRFC1123,
		Format:  Hostname,
	},
	"fqdn": {
		Pattern: FqdnRegexStringRFC1123,
		Format:  Hostname,
	},
	"btc_addr":                  {Pattern: BtcAddressRegexString},
	"btc_addr_bech32":           {Pattern: BtcAddressBech32RegexString},
	"eth_addr":                  {Pattern: EthAddressRegexString},
	"url_encoded":               {Pattern: URLEncodedRegexString},
	"html":                      {Pattern: HTMLRegexString},
	"html_encoded":              {Pattern: HTMLEncodedRegexString},
	"jwt":                       {Pattern: JWTRegexString},
	"bic":                       {Pattern: BicRegexString},
	"semver":                    {Pattern: SemverRegexString},
	"dns_rfc1035_label":         {Pattern: DNSRegexStringRFC1035Label},
	"cve":                       {Pattern: CVERegexString},
	"mongodb":                   {Pattern: MongodbIdRegexString},
	"mongodb_connection_string": {Pattern: MongodbConnStringRegexString},
	"cron":                      {Pattern: CronRegexString},
	"spicedb":                   {Pattern: SpicedbIDRegexString},
	"ein":                       {Pattern: EinRegexString},

	"cidr":          {Pattern: CIDRRegex},
	"cidrv4":        {Pattern: CIDRRegex},
	"hostname_port": {Format: Hostname},
	"ip":            {Pattern: IPAnyRegexString},
	"ip_addr":       {Pattern: IPAnyRegexString},
	"ipv4":          {Format: IPv4},
	"ip4_addr":      {Format: IPv4},
	"ipv6":          {Format: IPv6},
	"ip6_addr":      {Format: IPv6},
	"mac":           {Pattern: MACRegexString},
	"tcp4_addr":     {Pattern: TCP4AddrRegexString},
	"tcp6_addr":     {Pattern: TCP6AddrRegexString},
	"tcp_addr":      {Pattern: TCPAddrRegexString},
	"udp4_addr":     {Pattern: UDP4AddrRegexString},
	"udp6_addr":     {Pattern: UDP6AddrRegexString},
	"udp_addr":      {Pattern: UDPAddrRegexString},
	"unix_addr":     {Pattern: UnixAddrRegexString},
	"uri":           {Format: URI},
	"url":           {Format: URI},
	"urn_rfc2141":   {Pattern: URNRFC2141RegexString},

	"required": {Required: true},
}
