package officialAccount

import (
	"context"
	"fmt"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/contract"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
)

// SendText 发送客服文本消息到指定用户 openID
// 返回 SDK 原始响应对象，通常为 *response.ResponseOfficialAccount
func (o *OfficalAccount) SendText(ctx context.Context, openID string, content string) (interface{}, error) {
	if openID == "" {
		return nil, fmt.Errorf("openID 不能为空")
	}
	if content == "" {
		return nil, fmt.Errorf("content 不能为空")
	}
	msg := messages.NewText(content)
	return o.app.CustomerService.Message(ctx, msg).SetTo(openID).Send(ctx)
}

// SendMessage 发送任意客服消息（文本、图片、卡片等），直接使用 PowerWeChat 的消息构造器
// 示例：msg := messages.NewImage(mediaID)
func (o *OfficalAccount) SendMessage(ctx context.Context, openID string, msg contract.MessageInterface) (interface{}, error) {
	if openID == "" {
		return nil, fmt.Errorf("openID 不能为空")
	}
	if msg == nil {
		return nil, fmt.Errorf("msg 不能为空")
	}
	return o.app.CustomerService.Message(ctx, msg).SetTo(openID).Send(ctx)
}
