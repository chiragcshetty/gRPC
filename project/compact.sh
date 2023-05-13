# use: bash compact.sh <hfiles_path> <new_file_name>

cd $1
sort -m -V -o merged.log *.dat
rm *.dat
uniq -w 4 merged.log > "$2.dat"
rm merged.log 

cd -