package prompttypes

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shank318/doota/ai"
	"io"
	"strings"

	"github.com/streamingfast/dstore"
	"go.uber.org/zap"
)

type Reader struct {
	store  dstore.Store
	logger *zap.Logger
}

func NewReader(messageTypesFolderPath string, logger *zap.Logger) (*Reader, error) {
	store, err := dstore.NewStore(messageTypesFolderPath, "", "", false)
	if err != nil {
		return nil, err
	}

	return &Reader{
		store:  store,
		logger: logger,
	}, nil
}

func (r *Reader) Reader(ctx context.Context) (*Store, error) {
	messageTypes := map[string]*promptTypeFiles{}

	err := r.store.Walk(ctx, "", func(filename string) error {
		if strings.HasPrefix(filename, "_") {
			r.logger.Debug("skipping file", zap.String("filename", filename))
			return nil
		}

		// filename can be: QUOTE_REQUEST/prompt.gotmpl OR PEREGRINE_LOAD_CREATION/child_prompt/human.gotmpl

		chunks := strings.Split(filename, "/")
		if len(chunks) < 2 || len(chunks) > 3 {
			r.logger.Warn("skipping invalid filename", zap.String("filename", filename))
			return nil
		}

		r.logger.Warn("filename", zap.String("filename", filename))

		mt, found := messageTypes[chunks[0]]
		if !found {
			mt = &promptTypeFiles{name: chunks[0]}
			messageTypes[chunks[0]] = mt
		}

		promptFileName := chunks[1]

		content := r.mustReadFile(ctx, filename)

		switch promptFileName {
		case "info.json":
			info := &basicInfo{}
			if err := json.Unmarshal(content, info); err != nil {
				return fmt.Errorf("unmarshal basicInfo %s: %w", filename, err)
			}
			mt.description = info.Description
			mt.getPromptConfig().Model = ai.GPTModel(info.Model)
		case "human.gotmpl":
			mt.getPromptConfig().HumanTmpl = string(content)
		case "prompt.gotmpl":
			mt.getPromptConfig().PromptTmpl = string(content)
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("walk: %w", err)
	}

	var out []*promptTypeFiles
	for k, mt := range messageTypes {
		if err := mt.validate(); err != nil {
			return nil, fmt.Errorf("invalid message type %s: %w", k, err)
		}
		out = append(out, mt)
	}

	return newStore(out), nil
}

func (r *Reader) mustReadFile(ctx context.Context, filename string) []byte {
	reader, err := r.store.OpenObject(ctx, filename)
	if err != nil {
		panic(fmt.Errorf("open file %s: %w", filename, err))
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("read buffer %s: %w", filename, err))
	}
	return data
}
