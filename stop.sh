#!/bin/bash
ps -ef | grep chia_monitor | grep -v grep | awk '{print $2}' | xargs kill -9
