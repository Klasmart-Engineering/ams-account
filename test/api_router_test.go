package test_test

import (
	"testing"
)

func TestAPIRouter(t *testing.T) {
	// testsetup.Setup()

	// api, err := openapi3.Load(apiDefinitionPath)
	// if err != nil {
	// 	panic(err)
	// }

	// backupLogger := logger.GetLogger()
	// logger.SetLogger(nil)

	// rootRouter := apirouter.NewRouter()
	// if globals.CORSOptions != nil {
	// 	rootRouter.AddCORSMiddleware(globals.CORSOptions)
	// }

	// routerV1 := handlers.CreateLambdaRouterV1()
	// rootRouter.AddRouter("v1", routerV1)

	// openapi3.TestRouter(t, api, rootRouter, &openapi3.RouterTestingOptions{
	// 	BasePath:        "/v1/",
	// 	IgnoreResources: []string{},
	// })

	// logger.SetLogger(backupLogger)
}

func TestAPIV2Router(t *testing.T) {
	// testsetup.Setup()

	// api, err := openapi3.Load(apiDefinitionPath_V2)
	// if err != nil {
	// 	panic(err)
	// }

	// backupLogger := logger.GetLogger()
	// logger.SetLogger(nil)

	// rootRouter := apirouter.NewRouter()
	// if globals.CORSOptions != nil {
	// 	rootRouter.AddCORSMiddleware(globals.CORSOptions)
	// }

	// routerV1 := handlers.CreateRouterV2()
	// rootRouter.AddRouter("v2", routerV1)

	// openapi3.TestRouter(t, api, rootRouter, &openapi3.RouterTestingOptions{
	// 	BasePath:        "/v2/",
	// 	IgnoreResources: []string{},
	// })

	// logger.SetLogger(backupLogger)
}
