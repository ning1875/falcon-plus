#!/bin/sh
cd  ~/go/src/github.com/open-falcon/falcon-plus/bin/transfer/

kill -9 $(pgrep falcon-transfer)
#ps -ef |grep falcon-transfer |awk '{print $2}'|xargs kill -9
mv -b falcon-transfer falcon-transfer.old
mv -b  falcon-transfer.new  falcon-transfer
echo "This is new"
./falcon-transfer &
sleep 30s
kill -9 $(pgrep falcon-transfer)
mv -b  falcon-transfer falcon-transfer.new
mv -b  falcon-transfer.old  falcon-transfer
echo "Rollback"
./falcon-transfer &

