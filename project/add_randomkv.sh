## Usage: add_kv.sh <filename> <KEY> <size_of_value_bytes>
echo -n "+$2 " >> $1
base64 /dev/urandom | head -c $3 >> $1
echo "" >> $1 
