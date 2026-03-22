// padoval is an approach for serving a HTTP API based on the principle of "Parse, don't validate".
//
// Padoval looks like regular go http handlers, but instead of passing a unparsed request,
// we expect a parsed request. We can do this because we are aiming at serving a opinionated REST API,
// not a general purpose whatever-http-can-do server.
package padoval
