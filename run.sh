set -e
set -o pipefail

./eggshellver migrate
export TESTING=true
go test -v ./... -timeout 120s
export TESTING=false
./eggshellver createcache
./eggshellver start