# e-transcriber

A minimal WebRTC audio ingest service using Pion and `go-whisper` that transcribes incoming audio in real time.

## Features

- `/session` endpoint for WebRTC SDP offers.
- Transcribes inbound audio with `go-whisper` and prints segments to stdout.
- Static demo page at `/demo.html`.

## Requirements

- Go 1.21+
- A Whisper model file in `models/` (downloaded automatically on first run).

## Running locally

```bash
mkdir -p models

go run .
```

The server will download `ggml-tiny.bin` into `models/` on first start.

Visit <http://localhost:8080/demo.html> and allow microphone access. Transcriptions appear in the server logs.

### Environment variables

| Variable | Default | Description |
| --- | --- | --- |
| `ADDR` | `:8080` | HTTP listen address |
| `WHISPER_MODELS_DIR` | `models` | Directory for Whisper models |
| `WHISPER_MODEL` | `ggml-tiny` | Model ID to use if already downloaded |
| `WHISPER_MODEL_PATH` | `ggml-tiny.bin` | Model filename or URL to download |
| `TRANSCRIBE_WINDOW_SECS` | `5` | Chunk duration in seconds |

## Docker

```bash
docker build -t e-transcriber .

docker run --rm -p 8080:8080 -v $(pwd)/models:/app/models e-transcriber
```

## Tests

```bash
go test ./...
```
