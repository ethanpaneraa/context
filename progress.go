package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/schollz/progressbar/v2"
)

type ProgressTracker struct {
    bar       *progressbar.ProgressBar
    total     int64
    current   int64
    mu        sync.Mutex
    startTime time.Time
}

func NewProgressTracker(total int64, description string) *ProgressTracker {
    bar := progressbar.NewOptions64(
        total,
        progressbar.OptionSetDescription(description),
        progressbar.OptionSetWidth(40),
        progressbar.OptionShowCount(),
        progressbar.OptionSetTheme(progressbar.Theme{
            Saucer:        "=",
            SaucerHead:    ">",
            SaucerPadding: " ",
            BarStart:      "[",
            BarEnd:        "]",
        }),
    )

    return &ProgressTracker{
        bar:       bar,
        total:     total,
        current:   0,
        startTime: time.Now(),
    }
}

func (pt *ProgressTracker) Increment(n int64) {
    pt.mu.Lock()
    defer pt.mu.Unlock()
    
    pt.current += n
    _ = pt.bar.Add64(n)
}

func (pt *ProgressTracker) IncrementFiles(count int) {
    pt.mu.Lock()
    defer pt.mu.Unlock()
    
    for i := 0; i < count; i++ {
        _ = pt.bar.Add(1)
    }
}

func (pt *ProgressTracker) Finish() {
    pt.mu.Lock()
    defer pt.mu.Unlock()
    
    _ = pt.bar.Finish()
    duration := time.Since(pt.startTime).Round(time.Millisecond)
    fmt.Printf("\nCompleted in %v\n", duration)
}