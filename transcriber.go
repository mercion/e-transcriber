package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
)

const whisperSampleRate = int(whisper.SampleRate)

const (
	defaultWindowSeconds = 5
)

type Transcriber struct {
	manager       *whisper.Manager
	model         *schema.Model
	windowSamples int

	ctx    context.Context
	cancel context.CancelFunc
	input  chan []float32
	wg     sync.WaitGroup
	mu     sync.Mutex
	offset time.Duration
}

func NewTranscriber(manager *whisper.Manager, model *schema.Model, windowSeconds int) *Transcriber {
	if windowSeconds <= 0 {
		windowSeconds = defaultWindowSeconds
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Transcriber{
		manager:       manager,
		model:         model,
		windowSamples: windowSeconds * whisperSampleRate,
		ctx:           ctx,
		cancel:        cancel,
		input:         make(chan []float32, 8),
	}
}

func (t *Transcriber) Start() {
	t.wg.Add(1)
	go t.run()
}

func (t *Transcriber) Stop() {
	t.cancel()
	t.wg.Wait()
}

func (t *Transcriber) Push(samples []float32) {
	select {
	case t.input <- samples:
	case <-t.ctx.Done():
	}
}

func (t *Transcriber) run() {
	defer t.wg.Done()
	buffer := make([]float32, 0, t.windowSamples*2)

	for {
		select {
		case <-t.ctx.Done():
			return
		case samples := <-t.input:
			buffer = append(buffer, samples...)
			for len(buffer) >= t.windowSamples {
				chunk := append([]float32(nil), buffer[:t.windowSamples]...)
				buffer = buffer[t.windowSamples:]
				if err := t.transcribeChunk(chunk); err != nil {
					fmt.Fprintf(os.Stderr, "transcription error: %v\n", err)
				}
				if t.ctx.Err() != nil {
					return
				}
			}
		}
	}
}

func (t *Transcriber) transcribeChunk(samples []float32) error {
	t.mu.Lock()
	offset := t.offset
	t.offset += time.Duration(len(samples)) * time.Second / time.Duration(whisperSampleRate)
	t.mu.Unlock()

	ctx, cancel := context.WithTimeout(t.ctx, 30*time.Second)
	defer cancel()

	return t.manager.WithModel(t.model, func(task *whisper.Task) error {
		return task.Transcribe(ctx, offset, samples, func(segment *schema.Segment) {
			segment.WriteText(os.Stdout)
		})
	})
}

func downsampleBy3(input []float32) []float32 {
	if len(input) < 3 {
		return nil
	}
	output := make([]float32, len(input)/3)
	for i := range output {
		idx := i * 3
		output[i] = (input[idx] + input[idx+1] + input[idx+2]) / 3
	}
	return output
}
