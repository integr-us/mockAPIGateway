package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

func splitSingleAndMultiValue(v map[string][]string) (single map[string]string, multi map[string][]string) {
	single = make(map[string]string)
	multi = make(map[string][]string)

	for key, val := range v {
		if len(v) > 1 {
			multi[key] = val
		} else {
			single[key] = val[0]
		}
	}
	return
}

func (app *application) dispatchEventHandler(w http.ResponseWriter, r *http.Request) {
	headers, mvHeaders := splitSingleAndMultiValue(r.Header)
	queryParams, mvQueryParams := splitSingleAndMultiValue(r.URL.Query())

	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.logger.Error("reading body", zap.Error(err))
	}

	payload := &events.APIGatewayProxyRequest{
		Resource:                        "",
		Path:                            r.URL.Path,
		HTTPMethod:                      r.Method,
		Headers:                         headers,
		MultiValueHeaders:               mvHeaders,
		QueryStringParameters:           queryParams,
		MultiValueQueryStringParameters: mvQueryParams,
		PathParameters:                  nil,
		StageVariables:                  nil,
		RequestContext: events.APIGatewayProxyRequestContext{
			AccountID:         "1234567890",
			ResourceID:        "",
			OperationName:     "",
			Stage:             "",
			DomainName:        "",
			DomainPrefix:      "",
			RequestID:         "",
			ExtendedRequestID: "",
			Protocol:          "",
			Identity:          events.APIGatewayRequestIdentity{},
			ResourcePath:      "",
			Path:              "",
			Authorizer:        nil,
			HTTPMethod:        "",
			RequestTime:       "",
			RequestTimeEpoch:  0,
			APIID:             "",
		},
		Body:            string(body),
		IsBase64Encoded: false,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		app.logger.Error("marshaling payload", zap.Error(err))
	}

	deadline := time.Now().Add(15 * time.Second)
	request := messages.InvokeRequest{
		Payload:      jsonPayload,
		RequestId:    "",
		XAmznTraceId: "",
		Deadline: messages.InvokeRequest_Timestamp{
			Seconds: deadline.Unix(),
			Nanos:   deadline.UnixNano(),
		},
		InvokedFunctionArn:    "",
		CognitoIdentityId:     "",
		CognitoIdentityPoolId: "",
		ClientContext:         nil,
	}

	var response messages.InvokeResponse
	err = app.rpcClient.Call("Function.Invoke", request, &response)
	if err != nil {
		app.logger.Error("calling RPC client", zap.Error(err))
	}
	if response.Error != nil {
		app.logger.Error("from lambda", zap.Error(response.Error))
	}

	responsePayload := &events.APIGatewayProxyResponse{}
	err = json.Unmarshal(response.Payload, responsePayload)

	w.Write([]byte(responsePayload.Body))
}

func (app *application) route() *chi.Mux {
	router := chi.NewRouter()
	router.HandleFunc(fmt.Sprintf("%s*", app.config.contextPath), app.dispatchEventHandler)

	return router
}

func (app *application) serve() {
	srv := &http.Server{
		Addr:                         fmt.Sprintf("%s:%d", app.config.host, app.config.port),
		Handler:                      app.route(),
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  0,
		ReadHeaderTimeout:            0,
		WriteTimeout:                 0,
		IdleTimeout:                  0,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	app.logger.Info(
		"starting server",
		zap.String("host", app.config.host),
		zap.Int("port", app.config.port),
		zap.String("context_path", app.config.contextPath),
	)

	err := srv.ListenAndServe()
	if err != nil {
		app.logger.Fatal("starting server", zap.Error(err))
	}
}
