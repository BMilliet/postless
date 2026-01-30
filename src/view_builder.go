package src

type ViewBuilderInterface interface {
	NewListView(title string, op []ListItem, height int) ListItem
	NewTextFieldView(title, placeHolder string) string
	NewCollectionsView(collections []Collection, config *ConfigJSON, secret *SecretJSON, configLoader *ConfigLoader, fileManager FileManagerInterface) string
	NewBodyEditorView(body interface{}) string
	NewRequestPreviewView(selectedRequest *RequestItem, config *ConfigJSON, secret *SecretJSON, configLoader *ConfigLoader) string
}

type ViewBuilder struct{}

func NewViewBuilder() *ViewBuilder {
	return &ViewBuilder{}
}

func (b *ViewBuilder) NewListView(title string, op []ListItem, height int) ListItem {
	endValue := ListItem{}
	ListView(title, op, height, &endValue)
	return endValue
}

func (b *ViewBuilder) NewTextFieldView(title, placeHolder string) string {
	endValue := ""
	TextFieldView(title, placeHolder, &endValue)
	return endValue
}

func (b *ViewBuilder) NewCollectionsView(collections []Collection, config *ConfigJSON, secret *SecretJSON, configLoader *ConfigLoader, fileManager FileManagerInterface) string {
	selected := ""
	CollectionsView(collections, config, secret, configLoader, fileManager, &selected)
	return selected
}

func (b *ViewBuilder) NewBodyEditorView(body interface{}) string {
	selected := ""
	BodyEditorView(body, &selected)
	return selected
}

func (b *ViewBuilder) NewRequestPreviewView(selectedRequest *RequestItem, config *ConfigJSON, secret *SecretJSON, configLoader *ConfigLoader) string {
	action := ""
	RequestPreviewView(selectedRequest, config, secret, configLoader, &action)
	return action
}
