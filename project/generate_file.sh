## Usage: generate_file.sh <filename> <size_in_bytes>
base64 /dev/urandom | head -c $2 > $1