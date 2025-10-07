package ports

import "context"

type TemplateRender interface {
	Process(ctx context.Context, userEmail string, csvSourcePath string, templateHtml string) error
}