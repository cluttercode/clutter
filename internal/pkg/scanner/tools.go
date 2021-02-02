package scanner

import (
	"fmt"

	"go.uber.org/zap"
)

type toolFunc func(z *zap.SugaredLogger, path string, f func(*RawElement) error) error

var tools = map[string]func(ToolConfig) (toolFunc, error){
	"builtin": makeBuiltin,
}

func makeTool(cfg ToolConfig) (toolFunc, error) {
	mktool := tools[cfg.Tool]
	if mktool == nil {
		return nil, fmt.Errorf("unknown tool %s", cfg.Tool)
	}

	tool, err := mktool(cfg)
	if err != nil {
		return nil, fmt.Errorf("tool %s create: %w", cfg.Tool, err)
	}

	return tool, nil
}
