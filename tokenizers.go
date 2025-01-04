package main

import (
	"fmt"
	"path/filepath"

	"github.com/pkoukk/tiktoken-go"
	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
)



type Tokenizer interface {
	CountTokens(text string) (TokenCount, error)
	Name() string
}

type TiktokenTokenizer struct {
	encoding     *tiktoken.Tiktoken
	model       string
	tokenLimit  int
	warnLimit   int
}

func NewTiktokenTokenizer(model string, limit int) (*TiktokenTokenizer, error) {
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to get tiktoken encoding: %w", err)
	}

	warnLimit := int(float64(limit) * 0.8) 

	return &TiktokenTokenizer{
		encoding:    encoding,
		model:      model,
		tokenLimit: limit,
		warnLimit:  warnLimit,
	}, nil
}

func (t *TiktokenTokenizer) CountTokens(text string) (TokenCount, error) {
	tokens := t.encoding.Encode(text, nil, nil)
	count := len(tokens)
	
	tokenCount := TokenCount{
		Count:      count,
		TokensPerc: float64(count) / float64(t.tokenLimit) * 100,
		Truncated:  count > t.tokenLimit,
		TokenLimit: t.tokenLimit,
		WarnLimit:  t.warnLimit,
	}

	return tokenCount, nil
}

func (t *TiktokenTokenizer) Name() string {
	return fmt.Sprintf("tiktoken-%s", t.model)
}

type HuggingFaceTokenizer struct {
	tokenizer   *tokenizer.Tokenizer
	modelPath   string
	tokenLimit  int
	warnLimit   int
}

func NewHuggingFaceTokenizer(modelPath string, limit int) (*HuggingFaceTokenizer, error) {
	tok, err := pretrained.FromFile(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load HuggingFace tokenizer: %w", err)
	}

	warnLimit := int(float64(limit) * 0.8)

	return &HuggingFaceTokenizer{
		tokenizer:  tok,
		modelPath:  modelPath,
		tokenLimit: limit,
		warnLimit:  warnLimit,
	}, nil
}

func (h *HuggingFaceTokenizer) CountTokens(text string) (TokenCount, error) {
	encoding, err := h.tokenizer.EncodeSingle(text)
	if err != nil {
		return TokenCount{}, fmt.Errorf("failed to encode text: %w", err)
	}

	count := len(encoding.Ids)
	
	tokenCount := TokenCount{
		Count:      count,
		TokensPerc: float64(count) / float64(h.tokenLimit) * 100,
		Truncated:  count > h.tokenLimit,
		TokenLimit: h.tokenLimit,
		WarnLimit:  h.warnLimit,
	}

	return tokenCount, nil
}

func (h *HuggingFaceTokenizer) Name() string {
	return fmt.Sprintf("huggingface-%s", filepath.Base(h.modelPath))
}

func NewTokenizer(tokType TokenizerType, modelPath string, limit int) (Tokenizer, error) {
	switch tokType {
	case TiktokenGPT35:
		return NewTiktokenTokenizer("gpt-3.5-turbo", limit)
	case TiktokenGPT4:
		return NewTiktokenTokenizer("gpt-4", limit)
	case TiktokenClaude:
		return NewTiktokenTokenizer("claude", limit)
	case HuggingFace:
		return NewHuggingFaceTokenizer(modelPath, limit)
	default:
		return nil, fmt.Errorf("unsupported tokenizer type: %s", tokType)
	}
}