# API package

This is the place for actual "controllers" from mvc. The place where "the outside" makes requests to our service. Be it HTTP, be it gRPC, be it RPC. Maybe some Domain transfer objects (DTO) between this package and the domain package are needed in the future. 

NOTE: Currently this package is empty. The server is initialized in app.go using the domain layers handler package directly, but handler is a domain service, it should not know about http at all. the handler should be called using this api package instead following strict DDD, then the API package calls a func like "domain.GetHello(string param)". Not yet implemented, as I focused on server abstraction to make the principle clear.