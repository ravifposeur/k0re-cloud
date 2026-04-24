#!/bin/bash
# scripts/modules/04_lifecycle.sh

ACTION=$1
NAME=$2

case "$ACTION" in
    "start")
        echo "  -> [Module: lifecycle] Menyalakan kembali container $NAME..."
        docker start "$NAME" >/dev/null
        echo "[$(date +"%Y-%m-%d %H:%M:%S")] ACTION=START SERVER=$NAME STATUS=SUCCESS" >> "./k0re-data/audit.log"
        ;;
    "stop")
        echo "  -> [Module: lifecycle] Menghentikan container $NAME..."
        docker stop "$NAME" >/dev/null
        echo "[$(date +"%Y-%m-%d %H:%M:%S")] ACTION=STOP SERVER=$NAME STATUS=SUCCESS" >> "./k0re-data/audit.log"
        ;;
    "destroy")
        echo "  -> [Module: lifecycle] Menghancurkan container $NAME secara permanen..."
        docker rm -f "$NAME" >/dev/null
        rm -f "/tmp/k0re_opts_$NAME" "/tmp/k0re_dir_$NAME"
        echo "[$(date +"%Y-%m-%d %H:%M:%S")] ACTION=DESTROY SERVER=$NAME STATUS=SUCCESS" >> "./k0re-data/audit.log"
        ;;
    *)
        echo "Aksi tidak dikenal: $ACTION"
        exit 1
        ;;
esac