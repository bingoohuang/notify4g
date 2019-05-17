#!/usr/bin/env bash

set -x #echo on

# search css/js on https://www.bootcdn.cn/.

mkdir -p res/bootstrap/3.3.1/css
(cd res/bootstrap/3.3.1/css && curl -LO https://cdn.bootcss.com/twitter-bootstrap/3.3.1/css/bootstrap.min.css)
mkdir -p res/codemirror/5.45.0
(cd res/codemirror/5.45.0 && curl -LO https://cdn.bootcss.com/codemirror/5.45.0/codemirror.min.css)
(cd res/codemirror/5.45.0 && curl -LO https://cdn.bootcss.com/codemirror/5.45.0/codemirror.min.js)
(cd res/codemirror/5.45.0 && curl -LO https://cdn.bootcss.com/codemirror/5.45.0/mode/javascript/javascript.min.js)
(cd res/codemirror/5.45.0 && curl -LO https://cdn.bootcss.com/codemirror/5.46.0/mode/meta.min.js)
(cd res/codemirror/5.45.0 && curl -LO https://cdn.bootcss.com/codemirror/5.46.0/addon/mode/loadmode.min.js)
mkdir -p res/jquery/2.1.3
(cd res/jquery/2.1.3 && curl -LO https://cdn.bootcss.com/jquery/2.1.3/jquery.min.js)
go get github.com/bingoohuang/statiq
statiq -src=res
#rm -fr res/bootstrap
#rm -fr res/codemirror
#rm -fr res/jquery
