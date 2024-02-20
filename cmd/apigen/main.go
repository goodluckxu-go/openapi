package main

import (
	"fmt"
	"github.com/goodluckxu-go/openapi"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultRootDir  = "./"
	defaultRouteDir = defaultRootDir
	defaultDocPath  = defaultRootDir + "doc.go"
	defaultOutDir   = defaultRootDir + "docs"
)

func main() {
	app := cli.NewApp()
	app.Version = openapi.Version
	app.Usage = "使用openapi自动生成RESTFULL API文档。"
	app.Commands = []*cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "创建 openapi.json和openapi.yaml 文档",
			Action: func(ctx *cli.Context) error {
				rootDir, _ := ctx.Value("rootDir").(string)
				if rootDir == "" {
					rootDir = defaultRootDir
				}
				routeDir, _ := ctx.Value("routeDir").(string)
				if routeDir == "" {
					routeDir = defaultRouteDir
				}
				docPath, _ := ctx.Value("docPath").(string)
				if docPath == "" {
					docPath = defaultDocPath
				}
				outDir, _ := ctx.Value("outDir").(string)
				if outDir == "" {
					outDir = defaultOutDir
				}
				ginRouteDir, _ := ctx.Value("ginRouteDir").(string)
				openapi.GenerateOpenAPI(rootDir, routeDir, docPath, outDir, ginRouteDir)
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "rootDir",
					Usage:       "项目根目录",
					DefaultText: defaultRootDir,
				},
				&cli.StringFlag{
					Name:        "routeDir",
					Usage:       "路由注释目录",
					DefaultText: defaultRouteDir,
				},
				&cli.StringFlag{
					Name:        "docPath",
					Usage:       "文档注释地址",
					DefaultText: defaultDocPath,
				},
				&cli.StringFlag{
					Name:        "outDir",
					Usage:       "生成文档输出目录",
					DefaultText: defaultOutDir,
				},
				&cli.StringFlag{
					Name:  "ginRouteDir",
					Usage: "gin生成路由文件",
				},
			},
		},
		{
			Name:    "downSwagger",
			Aliases: []string{"d"},
			Usage:   "下载 最新的swagger-ui 文件",
			Action: func(ctx *cli.Context) error {
				outDir, _ := ctx.Value("outDir").(string)
				if outDir == "" {
					outDir = defaultOutDir
				}
				fileInfo, err := os.Stat(outDir)
				if err != nil {
					_ = os.MkdirAll(outDir, 0777)
				} else if !fileInfo.IsDir() {
					return fmt.Errorf("已存在 %v 文件", outDir)
				}
				// 不存在压缩包则下载
				filePath := filepath.Join(outDir, "swagger.tar.gz")
				downUrl := ""
				if !openapi.IsFile(filePath) {
					downUrl, err = openapi.GetGithubLastDownUrl("swagger-api/swagger-ui")
					if err != nil {
						return err
					}
					if err = openapi.Download(downUrl, filePath); err != nil {
						return err
					}
				}
				return openapi.UnSwaggerTarball(filePath, outDir)
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "outDir",
					Usage:       "下载文件地址",
					DefaultText: defaultOutDir,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
