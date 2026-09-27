echo "=== step 2 ==="

echo "1. Please, upload the file './data/new_peoplenyc1080p_new.mp4' to Youtube"
echo "2. Wait about 10 minutes while Youtube chews the file..."
echo "3. Download the chewed file from Youtube and move it to folder './data'"

# yt-dlp -f "bestvideo[height=1080]+bestaudio/best" -o "./data/%(title)s [%(id)s].%(ext)s" https://youtu.be/VIDEO_ID
