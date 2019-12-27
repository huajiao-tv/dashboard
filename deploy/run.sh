#!/bin/bash

/usr/local/nginx/sbin/nginx
/data/dashboard/bin/dashboard -e=test -cfg=/etcd/dashboard.yaml