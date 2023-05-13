# Finds a key and returns the next 20 lines
# Syntax findKey.sh "KEY" "FILENAME"
grep -A20 "$1 " $2