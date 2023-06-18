ffmpeg -i "peoplenyc1080p.mp4" -i "orig_10.mp3" -map 0:v -map 1:a -c:v copy -c:a copy "new_peoplenyc1080p.mp4" -y
