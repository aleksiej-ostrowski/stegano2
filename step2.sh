echo "=== step 2 ==="

echo "Attention! Steganography takes about 13Gb on SSD in this case"

# time ./stegano2 e 1 "123" "./data/dubrowskij.txt.crp" "./data/new_peoplenyc1080p.mp4"
time ./stegano2 e 1 "123" "./data/Jose.Raul.Capablanca2.mp4.crp" "./data/chai.mp4"
