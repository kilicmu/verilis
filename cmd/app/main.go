package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/user/go-scaffold/internal/command"
	"github.com/user/go-scaffold/pkg/config"
)

// Variables that will be populated at build time via ldflags
var (
	Version   = "0.1.0"
	BuildTime = "unknown"
	GitCommit = "unknown"

	// Environment variables
	Env     = "development"
	ApiURL  = "http://localhost:8080"
	Debug   = "false"
	AppName = ""
)

func main() {

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// 定义全局标志
	helpFlag := flag.Bool("help", false, "Show help information")

	// 解析命令行参数
	flag.Parse()

	// 如果请求显示帮助信息或没有提供子命令
	if *helpFlag || len(flag.Args()) == 0 {
		printHelp()
		return
	}

	// 获取子命令和参数
	subCommand := flag.Arg(0)
	args := flag.Args()[1:]

	// 创建命令执行器
	executor := command.NewExecutor(cfg)

	// 执行对应的子命令
	switch strings.ToLower(subCommand) {
	case "init":
		executor.Init(args)
	case "generate":
		executor.Generate(args)
	default:
		fmt.Fprintf(os.Stderr, "\033[31mUnknown command: %s\033[0m\nRun '%s --help' for usage information\n", subCommand, AppName)
		os.Exit(1)
	}
}

// 打印帮助信息
func printHelp() {
	fmt.Printf("\033[32m 「%s」 is an Ai-driven i18n solution\033[0m \n", AppName)
	fmt.Println("Usage:")
	fmt.Printf("  %s [global option] <command> [arguments...]\n\n", AppName)

	fmt.Println("Commands:")
	fmt.Println("init - initial project in this project")
	fmt.Println("generate - generate language resource")
	fmt.Println("")

	fmt.Println("Global options:")
	flag.PrintDefaults()
}
