# rm -rf ./pics
# mkdir ./pics
# rm ./log_qr.log 
# bash format.sh  

# new_peoplenyc1080p.mp4  p5.txt

# GO111MODULE=off go build stegano2.go && ./stegano2 e 1 "123" 1.bin test.webm

# GO111MODULE=off go build stegano2.go

# GO111MODULE=off go build stegano2.go && ./stegano2 d 1 "123" "1.bin_filesize20086400.webm"

# ./stegano2 d 1 "1111111111111111" "1.bin_filesize502160.webm"

# GO111MODULE=off go build stegano2.go && ./stegano2 d 1 "1111111111111111" '0.4 1 bin filesize30129600 [k33tUWZ67Ow].mkv'

# md5sum ./1.bin
# md5sum ./"1.bin_filesize502160.webm.original"

# GO111MODULE=off go build stegano2.go && time ./stegano2 e 1 "1111111111111111" "./data/Jose.Raul.Capablanca.mp4" "./data/GINGERGREEN.mp4"
# GO111MODULE=off go build stegano2.go && ./stegano2 d 1 "1111111111111111" "./data/Jose Raul Capablanca mp4 filesize496184320 [lQavQsjjYlU].mkv"


# python3 ./wrap/codilla.py -input "./data/dubrowskij.txt" -output "./data/dubrowskij.txt.crp" -e

# GO111MODULE=off go build stegano2.go && time ./stegano2 e 1 "123" "./data/dubrowskij.txt.crp" "./data/new_peoplenyc1080p.mp4"
GO111MODULE=off go build stegano2.go && time ./stegano2 e 3 "123" "./data/dubrowskij.txt.crp" "./data/new_peoplenyc1080p.mp4"

# GO111MODULE=off go build stegano2.go && time ./stegano2 d 1 "123" "./data/dubrowskij.txt.crp_filesize58435200.webm"

# python3 ./wrap/codilla.py -input "./data/dubrowskij.txt.crp_filesize58435200.webm.original" -output "./data/dubrowskij.txt.result" -d

# md5sum "./data/dubrowskij.txt"
# md5sum "./data/dubrowskij.txt.result"

# GO111MODULE=off go build stegano2.go && time ./stegano2 d 1 "123" "./data/dubrowskij txt crp filesize58435200 [7WhQfMocbQQ].mkv"

# python3 ./wrap/codilla.py -input "./data/dubrowskij txt crp filesize58435200 [7WhQfMocbQQ].mkv.original" -output "./data/dubrowskij.txt.youtube" -d

# md5sum "./data/dubrowskij.txt"
# md5sum "./data/dubrowskij.txt.youtube"

# md5sum "./data/Jose.Raul.Capablanca.mp4_filesize12404608.webm.original"
# ./compare "./data/Jose.Raul.Capablanca.mp4" "./data/Jose.Raul.Capablanca.mp4_filesize9303456.webm.original"
# ./compare "./data/Jose.Raul.Capablanca.mp4" "./data/Jose Raul Capablanca mp4 filesize9303456 [6CvfQIuUFBc].mkv.original"
