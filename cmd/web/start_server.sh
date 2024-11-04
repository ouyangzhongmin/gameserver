go build -o go_gs_web main.go
nohup ./go_gs_web -c ../../configs/config.toml > data.log 2>&1 &
