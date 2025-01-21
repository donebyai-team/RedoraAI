package ai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"

	"go.uber.org/zap"
)

func (c *Client) debugTemplates(ctx context.Context, runId string, vars map[string]any, templates []*template.Template, logger *zap.Logger) {
	if c.debugFileStore == nil {
		return
	}

	for _, tmpl := range templates {
		buf := new(bytes.Buffer)
		err := tmpl.Execute(buf, vars)
		if err != nil {
			logger.Debug("failed to execute template", zap.Error(err), zap.String("tmpl", tmpl.Name()))
			continue
		}

		c.saveFile(ctx, runId, strings.TrimSuffix(tmpl.Name(), ".gotmpl"), buf, logger)
	}
}

func (c *Client) saveOutput(ctx context.Context, runId, outputName string, output []byte, logger *zap.Logger) {
	if c.debugFileStore == nil {
		return
	}

	buf := bytes.NewBuffer(output)

	filepath := fmt.Sprintf("%s/%s.json", runId, outputName)
	if err := c.debugFileStore.WriteObject(ctx, filepath, buf); err != nil {
		logger.Debug("failed to save output", zap.Error(err), zap.String("output_name", outputName))
		return
	}
}

func (c *Client) saveFile(ctx context.Context, runId string, name string, f io.Reader, logger *zap.Logger) {
	filepath := fmt.Sprintf("%s/%s.txt", runId, name)
	if err := c.debugFileStore.WriteObject(ctx, filepath, f); err != nil {
		logger.Debug("failed to save tmpl render", zap.Error(err), zap.String("tmpl", name))
		return
	}
	logger.Debug("saved file", zap.String("filepath", filepath))
}
