python codilla.py -input test.webm -output test.webm.crp -e
python codilla.py -input test.webm.crp -output test.webm.original -d
md5sum ./test.webm.original 
md5sum ./test.webm 
