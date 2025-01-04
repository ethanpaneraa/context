package main

type TokenizerType string

const (
    TiktokenGPT35  TokenizerType = "gpt-3.5-turbo"
    TiktokenGPT4   TokenizerType = "gpt-4"
    TiktokenClaude TokenizerType = "claude"
    HuggingFace    TokenizerType = "huggingface"
)

type Config struct {
    Path           string
    Include        []string
    Exclude        []string
    MaxSize        int64
    MaxDepth       int
    Output         string
    Threads        int
    Hidden         bool
    UseClip        bool
    Interactive    bool
    TokenizerType  TokenizerType
    TokenizerModel string
    TokenLimit     int
}

type TokenCount struct {
    Count      int
    TokensPerc float64 
    Truncated  bool
    TokenLimit int
    WarnLimit  int 
}

type FileEntry struct {
    Path       string
    Content    string
    Size       int64
    TokenCount *TokenCount
}