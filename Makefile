compile_grpc:
	protoc -I=internal/grpc_consumer -I=internal/InputStreamShard \
	  --go_out=. --go-grpc_out=. \
	  --go_opt=module=go_video_streamer \
	  --go-grpc_opt=module=go_video_streamer \
	  internal/grpc_consumer/VideoRCV.proto \
	  internal/InputStreamShard/InputStreamShard.proto
