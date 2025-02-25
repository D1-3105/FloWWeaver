package client_stream_creator

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"go_video_streamer/api/file_server"
	"go_video_streamer/config"
	"go_video_streamer/internal/base_rpc"
	"go_video_streamer/internal/cryptography"
	"go_video_streamer/internal/opencv_global_capture"
	"go_video_streamer/internal/video_streaming"
	"gocv.io/x/gocv"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"os"
	"sync"
)

type StreamMapping struct {
	channelMapping base_rpc.ChannelMappingService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type PresignedURLResponse struct {
	PresignedURL string `json:"presigned_url"`
}

const (
	HandleRestreamPresigned = "/restream/"
	HandlePresignUrl        = "/restream/presigned/"
)

var sm *StreamMapping
var apiPrefix string

func streamFromURL(newStream *ClientNewStreamRequest) (string, error) {
	sha := sha256.Sum256([]byte(newStream.Url))
	shaString := fmt.Sprintf("%x", sha)[:10]

	if _, ok := sm.channelMapping.GetStream(shaString); ok {
		return shaString, nil
	}

	vc, err := opencv_global_capture.NewVideoCapture(newStream.Url)
	if err != nil {
		return "", err
	}

	captureContext := video_streaming.CaptureContextFromCapture(
		vc,
		video_streaming.CaptureParams{
			FPS:    vc.Get(gocv.VideoCaptureFPS),
			Width:  int(vc.Get(gocv.VideoCaptureFrameWidth)),
			Height: int(vc.Get(gocv.VideoCaptureFrameHeight)),
		},
		shaString,
	)

	switch newStream.StreamT {
	case uint32(config.AnythingToHLS):
		sm.channelMapping.SetStream(
			shaString,
			&base_rpc.StreamQ{
				Mu:             &sync.Mutex{},
				CaptureContext: captureContext,
			},
		)
		go video_streaming.ListenStreamToHLS(captureContext, shaString)
		return shaString, nil
	default:
		return "", errors.New("stream not supported")
	}
}

func writeError(w http.ResponseWriter, errorString string) {
	w.WriteHeader(http.StatusBadRequest)
	errResp, _ := json.Marshal(ErrorResponse{Error: errorString})
	_, _ = w.Write(errResp)
}

// @Summary Generate presigned URL
// @Description Generates an encrypted presigned URL for creating a new stream
// @Tags Stream
// @Accept json
// @Produce json
// @Param request body ClientNewStreamRequest true "Stream configuration"
// @Success 200 {object} PresignedURLResponse
// @Failure 400 {object} ErrorResponse
// @Router /restream/presigned/ [post]
func presignedUrl(w http.ResponseWriter, r *http.Request) {
	var data ClientNewStreamRequest
	key, exists := os.LookupEnv("FLOWWEAVER_SECRET_KEY")
	if !exists {
		writeError(w, "Server has no secret key!")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if err := json.Unmarshal(body, &data); err != nil {
		writeError(w, err.Error())
		return
	}

	marshalled, err := proto.Marshal(&data)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	encrypted, err := cryptography.Encrypt([]byte(key), marshalled)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	fullURL := fmt.Sprintf(
		"%s://%s%s?signature=%s",
		scheme, r.Host, apiPrefix+HandleRestreamPresigned, base64.StdEncoding.EncodeToString(encrypted),
	)

	serialized, _ := json.Marshal(PresignedURLResponse{PresignedURL: fullURL})
	_, _ = w.Write(serialized)
}

// @Summary Process stream
// @Description Processes the presigned URL and starts the stream
// @Tags Stream
// @Produce json
// @Param signature query string true "Encrypted stream configuration"
// @Success 302
// @Failure 400 {object} ErrorResponse
// @Router /restream/ [get]
func makeStream(w http.ResponseWriter, r *http.Request) {
	key, exists := os.LookupEnv("FLOWWEAVER_SECRET_KEY")
	if !exists {
		writeError(w, "Server has no secret key!")
		return
	}

	signature := r.URL.Query().Get("signature")
	if signature == "" {
		writeError(w, "Signature parameter is missing")
		return
	}
	signatureDecoded, err := base64.StdEncoding.DecodeString(signature)
	decrypted, err := cryptography.Decrypt([]byte(key), signatureDecoded)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	var newStream ClientNewStreamRequest
	if err := proto.Unmarshal(decrypted, &newStream); err != nil {
		writeError(w, fmt.Sprintf("Error during deserialization: %s", err.Error()))
		return
	}

	streamName, err := streamFromURL(&newStream)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	http.Redirect(w, r, file_server.BuildHLSPlaylistURL(streamName), http.StatusFound)
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Router(prefix string) *mux.Router {
	apiPrefix = prefix
	sm = &StreamMapping{
		channelMapping: base_rpc.NewChannelMappingMap(),
	}

	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc(HandleRestreamPresigned, makeStream).Methods(http.MethodGet)
	router.HandleFunc(HandlePresignUrl, presignedUrl).Methods(http.MethodPost)
	router.HandleFunc("/health/", healthCheck).Methods(http.MethodGet)

	return router
}
