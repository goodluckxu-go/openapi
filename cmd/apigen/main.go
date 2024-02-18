package main

import (
	"github.com/goodluckxu-go/openapi"
	"github.com/urfave/cli/v2"
	"log"
	"os"
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
				openapi.GenerateOpenAPI(rootDir, routeDir, docPath, outDir)
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
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
