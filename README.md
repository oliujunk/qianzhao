# 编译

$env:CC="arm-linux-gnueabihf-gcc"
$env:CXX="arm-linux-gnueabihf-g++"
$env:GOOS="linux"
$env:GOARCH="arm"
$env:CGO_ENABLED="1"
$env:GOARM="7"
$env:GIN_MODE="release"
go build
