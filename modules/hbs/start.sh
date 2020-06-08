#!/bin/bash

for i in $(seq 1 10)
do
    nohup ./hbs > ./log/$i".txt" &
done