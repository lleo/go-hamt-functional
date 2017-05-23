#!/usr/bin/env zsh

repeat 10 do
       go test -run=xxx -bench=Hamt
done | ./mean-stddev-hamt.pl
