package openapi

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

func GenerateOpenAPI(rootDir, routeDir, docPath, outDir, ginGenerateRouteDir string) {
	modPathMap = modHandle{}
	var err error
	projectModName, err = modPathMap.load(rootDir)
	if err != nil {
		log.Fatal(err)
	}
	openapi := &openapiHandle{}
	openapi.load(routeDir, docPath)
	if !isDir(outDir) {
		err = os.MkdirAll(outDir, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
	var buf []byte
	// 生成yaml文档
	buf, err = yamlMarshal(openapi.t)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(outDir, "openapi.yaml"), buf, 0777)
	if err != nil {
		log.Fatal(err)
	}
	// 生成json文档
	buf, err = json.MarshalIndent(&openapi.t, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(outDir, "openapi.json"), buf, 0777)
	if err != nil {
		log.Fatal(err)
	}
	if ginGenerateRouteDir != "" {
		gins := ginHandle{}
		gins.load(openapi.routesFunc, ginGenerateRouteDir)
	}
}
