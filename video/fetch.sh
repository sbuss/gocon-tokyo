#!/bin/bash

for i in {0..163}; do
curl "https://stream01.durasite.net/crash.academy/goconf-a1_efUpcvz_zIuOuTQE.mp4/media_w950834278_$i.ts"   \
    -H 'Connection: keep-alive' \
    -H 'User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36' \
    -H 'Accept: */*' \
    -H 'Origin: https://crash.academy' \
    -H 'Sec-Fetch-Site: cross-site' \
    -H 'Sec-Fetch-Mode: cors' \
    -H 'Sec-Fetch-Dest: empty' \
    -H 'Referer: https://crash.academy/' \
    -H 'Accept-Language: en-US,en;q=0.9' \
    --compressed --insecure --output "media_w950834278_$i.ts"
done

cat media_w950834278_{0..163}.ts > all.ts
ffmpeg -i all.ts -acodec copy -vcodec copy all.mp4
