package service

import (
	"context"

	"agent/internal/config"
	"agent/internal/tool"
	"agent/pkg/logger"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
	"go.uber.org/zap"
)

// RunAgent 执行核心 AI 逻辑 (升级为 ReAct Agent)
func RunAgent(ctx context.Context, prompt string) (string, error) {
	cfg := config.AppConfig.LLMConfig

	// 1. 初始化大模型 (DeepSeek 完美兼容 OpenAI 接口)
	llm, err := openai.New(
		openai.WithModel(cfg.Model),
		openai.WithBaseURL(cfg.BaseURL),
		openai.WithToken(cfg.APIKey),
	)
	if err != nil {
		logger.Log.Error("初始化大模型失败", zap.Error(err))
		return "", err
	}

	// 2. 注册工具箱 (你可以无限往里面塞自定义工具)
	agentTools := []tools.Tool{
		tool.ZhihuScraperTool{}, // 你之前写的 Mock 工具
		tool.GithubSearchTool{}, // Github 查询工具
	}

	// 3. 初始化 Agent 执行器 (ZeroShotReactDescription 是最经典的 ReAct 模式)
	executor, err := agents.Initialize(
		llm,
		agentTools,
		agents.ZeroShotReactDescription,
		agents.WithMaxIterations(3), // 限制最多思考 3 轮，防止大模型陷入死循环烧钱
	)
	if err != nil {
		logger.Log.Error("初始化 Agent 失败", zap.Error(err))
		return "", err
	}

	logger.Log.Info("🤖 Agent 开始思考并拆解任务...", zap.String("指令", prompt))

	// 4. 运行 Agent
	result, err := chains.Run(ctx, executor, prompt)
	if err != nil {
		logger.Log.Error("Agent 执行异常", zap.Error(err))
		return "", err
	}

	return result, nil
}
