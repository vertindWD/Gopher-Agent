package tool

import (
	"agent/pkg/logger"
	"context"
	"fmt"

	"github.com/tmc/langchaingo/tools"
	"go.uber.org/zap"
)

// ZhihuScraperTool 定义一个知乎爬虫工具
type ZhihuScraperTool struct{}

// 确保 ZhihuScraperTool 实现了 tools.Tool 接口
var _ tools.Tool = ZhihuScraperTool{}

// Name 返回工具的英文名称（大模型底层是通过这个名字来区分工具的）
func (t ZhihuScraperTool) Name() string {
	return "zhihu_search_and_scrape"
}

// Description 极其重要！这是写给大模型看的“说明书”。大模型靠这段话来判断什么时候该用这个工具。
func (t ZhihuScraperTool) Description() string {
	return "当你需要获取最新的行业动态、知乎热帖、或者用户真实讨论时，请使用此工具。输入应该是你需要搜索的准确关键词。"
}

// Call 真正的执行逻辑，当大模型决定调用此工具时，会触发这个 Go 函数
func (t ZhihuScraperTool) Call(ctx context.Context, input string) (string, error) {
	logger.Log.Info("🔧 Agent 决定调用工具: ZhihuScraperTool", zap.String("关键词", input))

	// TODO: 这里未来可以接入 goquery 爬虫 或 真实的搜索引擎 API (如 SerpAPI)
	// 现在我们先用 Mock 数据让大模型的推理链路跑通

	mockResult := fmt.Sprintf(`
搜索关键词: %s
获取到的高赞回答内容：
1. Go 1.22 引入了全新的 math/rand/v2 库，并且终于修复了 for 循环变量捕获的史诗级坑。
2. 现在大厂后端开发非常看重高并发、Kafka 中间件经验以及 AI 工程化落地能力。
	`, input)

	return mockResult, nil
}
