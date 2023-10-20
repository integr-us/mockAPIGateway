# Notice

Under development, not ready for usage. Pull requests, issues, and feature requests are welcome.

# mockAPIGateway

Creates a basic web server that mocks an AWS API Gateway acting as a proxy to lambda.

# Usage

Run your Go-based lambda by exporting the env var `_LAMBDA_SERVER_PORT`, then for example `go run project/cmd/lambdaHandler`.
Your lambda should automatically start an RPC listener on the port you specified.

Now export any configuration needed for `mockapigateway`, for example: `export MOCK_GATEWAY_CONTEXT_PATH=/api/v1`

You should be able to use any web request tools to execute queries against your lambda.
Example: `curl -i http://localhost:3000/api/v1/hello/world`
