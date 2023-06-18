echo "=== make_a_comparison_video ==="

ffmpeg -i "dubrowskij txt crp filesize58435200 [7WhQfMocbQQ].mkv" -i "dubrowskij.txt.crp_filesize23374080.webm" -filter_complex hstack compare2.webm
