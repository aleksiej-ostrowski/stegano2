echo "=== step 1 ==="

time ./stegano2 merge --key "123" --mode aggressive \
    --data "./data/dubrowskij.txt" \
    --original "./data/new_peoplenyc1080p.mp4" \
    --result "./data/new_peoplenyc1080p_new.mp4"
