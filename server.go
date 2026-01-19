package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
	"github.com/pion/opus"
	"github.com/pion/webrtc/v4"
)

type Server struct {
	cfg         Config
	transcriber *TranscriberFactory
}

type SessionRequest struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

type SessionResponse struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

type TranscriberFactory struct {
	manager *whisper.Manager
	model   *schema.Model
	window  int
}

func NewServer(cfg Config, manager *whisper.Manager, model *schema.Model) *Server {
	return &Server{
		cfg:         cfg,
		transcriber: &TranscriberFactory{manager: manager, model: model, window: cfg.WindowSecs},
	}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/session", s.handleSession)
	mux.Handle("/", http.FileServer(http.Dir("web")))
	return mux
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request", http.StatusBadRequest)
		return
	}

	var req SessionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.SDP == "" {
		http.Error(w, "missing sdp", http.StatusBadRequest)
		return
	}

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("peer connection error: %v", err), http.StatusInternalServerError)
		return
	}

	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("peer connection state: %s", state.String())
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			_ = peerConnection.Close()
		}
	})

	_, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("transceiver error: %v", err), http.StatusInternalServerError)
		return
	}

	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		if track.Kind() != webrtc.RTPCodecTypeAudio {
			return
		}
		log.Printf("incoming audio track: %s", track.Codec().MimeType)
		go s.handleAudioTrack(track)
	})

	offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: req.SDP}
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		http.Error(w, fmt.Sprintf("set remote description: %v", err), http.StatusBadRequest)
		return
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("create answer: %v", err), http.StatusInternalServerError)
		return
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		http.Error(w, fmt.Sprintf("set local description: %v", err), http.StatusInternalServerError)
		return
	}

	<-gatherComplete
	localDesc := peerConnection.LocalDescription()
	if localDesc == nil {
		http.Error(w, "local description missing", http.StatusInternalServerError)
		return
	}

	resp := SessionResponse{SDP: localDesc.SDP, Type: localDesc.Type.String()}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleAudioTrack(track *webrtc.TrackRemote) {
	decoder := opus.NewDecoder()
	transcriber := s.transcriber.New()
	transcriber.Start()
	defer transcriber.Stop()

	pcm := make([]float32, 960)

	for {
		packet, _, err := track.ReadRTP()
		if err != nil {
			log.Printf("track read error: %v", err)
			return
		}

		_, isStereo, err := decoder.DecodeFloat32(packet.Payload, pcm)
		if err != nil {
			log.Printf("opus decode error: %v", err)
			continue
		}
		if isStereo {
			log.Printf("stereo opus packet received; downmixing to mono")
		}

		downsampled := downsampleBy3(pcm)
		if len(downsampled) == 0 {
			continue
		}
		transcriber.Push(downsampled)
	}
}

func (t *TranscriberFactory) New() *Transcriber {
	return NewTranscriber(t.manager, t.model, t.window)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) Run() error {
	server := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on %s", s.cfg.Addr)
	return server.ListenAndServe()
}
