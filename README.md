# FloWWeaver

## What can it do now?
- Stream management+frame streaming through kafka (Recommended)
- Stream management+frame streaming through grpc (If you have no broker)
- Frame streaming through RabbitMQ
- GRPC based ComfyUI node (needs update of proto compilation)
- Transfer different stream protocols into HLS (checkout `./html` for JavaScript)

## TODO:
- RabbitMQ stream manager
- S3/MinIO integration (transfer only urls through channel)
- Metadata storage outside processes
- Kubernetes integration
- Horizontal autoscaling
