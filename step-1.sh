echo "=== step -1 ==="

echo "1. Install Go 1.22 or newer"
echo "2. Install ffmpeg 5.1 or newer (with libx264 and aac)"

go version
ffmpeg -version | head -n 1
ffprobe -version | head -n 1
