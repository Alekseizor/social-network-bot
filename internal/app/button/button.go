package button

import (
	"context"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
	"main/internal/app/model"
)

type VkClient struct {
	b *params.MessagesSendBuilder
	k *object.MessagesKeyboard
}

func New(ctx context.Context) (*VkClient, error) {
	client := &VkClient{}

	return client, nil
}

// Keyboard create Keyboard for current state
func (client *VkClient) Keyboard(items *model.ResponseBody) *object.MessagesKeyboard {
	//State check
	//if state==startState do
	client.k.AddRow()
	client.k.AddTextButton("Гайд", "", "primary")
	client.k.AddTextButton("Информация обо мне", "", "primary")

	for i := 0; i < len(items.Items); i++ {
		client.k.AddTextButton(items.Items[i].Caption, "", "primary")
	}
	client.k.AddRow()
	client.k.AddCallbackButton("Назад", "", "negative")

	return client.k
}

func (client *VkClient) ButtonHandler(ctx context.Context) {
}
