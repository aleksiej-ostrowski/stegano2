echo "=== step 6 ==="

md5sum "./data/dubrowskij.txt" "./data/dubrowskij.txt.youtube" | md5sum --check
