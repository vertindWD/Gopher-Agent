package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"agent/pkg/logger"

	"github.com/tmc/langchaingo/tools"
	"go.uber.org/zap"
)

// GithubSearchTool 真实的外部 API 调用工具
type GithubSearchTool struct{}

var _ tools.Tool = GithubSearchTool{}

func (t GithubSearchTool) Name() string {
	return "github_repo_search"
}

func (t GithubSearchTool) Description() string {
	return "当你需要查找 Github 上的开源项目、代码库或者某个技术方向（如 Go, AI, IoT）的热门仓库时，使用此工具。输入应该是你要搜索的技术关键词（如：Golang web framework）。"
}

func (t GithubSearchTool) Call(ctx context.Context, input string) (string, error) {
	logger.Log.Info("🔧 Agent 决定调用真实工具: GithubSearchTool", zap.String("关键词", input))

	// 1. URL 编码，防止关键词里有空格导致请求失败
	query := url.QueryEscape(input)

	// 2. 调用 Github 真实的公开 API (按星数降序排列)
	apiURL := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=stars&order=desc&per_page=3", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// 3. 发起 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 Github API 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 4. 解析 JSON 并提取核心信息喂给大模型
	var result struct {
		Items []struct {
			Name        string `json:"name"`
			HtmlUrl     string `json:"html_url"`
			Description string `json:"description"`
			Stars       int    `json:"stargazers_count"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析 Github 数据失败: %w", err)
	}

	if len(result.Items) == 0 {
		return "没有找到相关的 Github 仓库", nil
	}

	// 5. 组装结果返回给大模型的大脑
	var finalResponse string
	for i, item := range result.Items {
		finalResponse += fmt.Sprintf("%d. 项目名: %s (⭐ %d)\n   链接: %s\n   描述: %s\n\n",
			i+1, item.Name, item.Stars, item.HtmlUrl, item.Description)
	}

	return finalResponse, nil
}
