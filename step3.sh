echo "=== step 3 ==="

echo "1. Please, upload the file-result, for example './data/dubrowskij.txt.crp_filesize58435200.webm' to Youtube"
echo "2. Wait about 10 minutes while Youtube chews the file..."
echo "3. Download the chewed file from Youtube and move it to folder './data'. This file may have a name like 'dubrowskij txt crp filesize58435200 [7WhQfMocbQQ].mkv'"

# yt-dlp -f "bestvideo[height=1080]+bestaudio/best" -o "./data/%(title)s [%(id)s].%(ext)s" https://youtu.be/7WhQfMocbQQ
yt-dlp -f "bestvideo[height=1080]+bestaudio/best" -o "./data/%(title)s [%(id)s].%(ext)s" https://youtu.be/T5uRqsCwWG0
