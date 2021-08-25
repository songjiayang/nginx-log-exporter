# !/bin/sh

set -e
if [ $# != 1 ]; then
  echo "Usage:"

  echo "./build.sh 0.1.0"
  exit 1
fi

# prepare folders
rm -rf nginx-log-exporter && mkdir nginx-log-exporter
cp -r test nginx-log-exporter/test
cp config.yml nginx-log-exporter

declare -a os=(
"darwin"
"freebsd" "freebsd" "freebsd"
"linux" "linux" "linux" "linux" "linux" "linux" "linux" "linux" "linux" "linux"
"netbsd" "netbsd" "netbsd"
"openbsd" "openbsd" "openbsd"
"windows" "windows")

declare -a arches=(
"amd64"
"386" "amd64" "arm"
"386" "amd64" "arm" "arm64" "ppc64" "ppc64le" "mips" "mipsle" "mips64" "mips64le"
"386" "amd64" "arm"
"386" "amd64" "arm"
"386" "amd64")

len=${#os[@]}
for (( i=1; i<${len}+1; i++ ));
do
  echo "--> building nginx-log-exporter_$1_${os[$i-1]}_${arches[$i-1]}"
  GOOS="${os[$i-1]}" GOARCH="${arches[$i-1]}" go build -a -o nginx-log-exporter/nginx-log-exporter main.go
  tar cvzf "pkg/nginx-log-exporter_$1_${os[$i-1]}_${arches[$i-1]}.tar.gz" nginx-log-exporter
done
