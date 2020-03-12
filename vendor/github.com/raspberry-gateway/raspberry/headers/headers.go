package headers

// Fields of request headers.
const (
	UserAgent               = "User-Agent"
	ContentType             = "Content-Type"
	ContentLength           = "Content-Length"
	Authorization           = "Authorization"
	ContentEncoding         = "Content-Encoding"
	Accept                  = "Accept"
	AcceptEncoding          = "Accept-Encoding"
	StrictTransportSecurity = "Strict-Transport-Security"
	CacheControl            = "Cache-Control"
	Pragma                  = "Pragma"
	Expires                 = "Expires"
	Connection              = "Connection"
	WWWAuthenticate         = "WWW-Authenticate"
)

// Definitions of request content.
const (
	RaspberryHookshot = "Raspberry-Hookshot"
	ApplicationJSON   = "application/json"
	ApplicationXML    = "application/xml"
)

// keys of request Context.
const (
	XRealIP                 = "X-Real-IP"
	XForwardFor             = "X-Forwarded-For"
	XAuthResult             = "X-Auth-Result"
	XSessionAlias           = "X-Session-Alias"
	XInitialURI             = "X-Initial-URI"
	XForwardProto           = "X-Forward-Proto"
	XContentTypeOptions     = "X-Content-Type-Options"
	XXSSProtection          = "X-XSS-Protection"
	XFrameOptions           = "X-Frame-Options"
	XRaspberryNodeID        = "x-raspberry-nodeid"
	XRaspberryNonce         = "x-raspberry-nonce"
	XRaspberryHostname      = "x-respberry-hostname"
	XGenerator              = "X-Generator"
	XRaspberryAuthorization = "X-Raspberry-Authorization"
)

// Gateway's custom response headers
const (
	XRateLimitLimit     = "X-RateLimit-Limit"
	XRateLimitRemaining = "X-RateLimit-Remaining"
	XRateLimitReset     = "X-RateLimit-Reset"
)

// HTTP Context standard keys
const (
	RemoteAddr = "remote_addr"
)
