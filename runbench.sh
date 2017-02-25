#!/usr/bin/env bash

echo "Builtin Map implementation"
go test -H -run=xxx -bench=Map | tee map.b
perl -pi -e 's/Map//' map.b

echo "Hamt32 Hybrid/Compressed/Full"
go test -H -run=xxx -bench=Hamt32 | tee hamt32-hybrid.b
go test -C -run=xxx -bench=Hamt32 | tee hamt32-comp.b
go test -F -run=xxx -bench=Hamt32 | tee hamt32-full.b

perl -pi -e 's/Hamt32//' hamt32-hybrid.b
perl -pi -e 's/Hamt32//' hamt32-comp.b
perl -pi -e 's/Hamt32//' hamt32-full.b

echo "Hamt64 Hybrid/Compressed/Full"
go test -H -run=xxx -bench=Hamt64 | tee hamt64-hybrid.b
go test -C -run=xxx -bench=Hamt64 | tee hamt64-comp.b
go test -F -run=xxx -bench=Hamt64 | tee hamt64-full.b

perl -pi -e 's/Hamt64//' hamt64-hybrid.b
perl -pi -e 's/Hamt64//' hamt64-comp.b
perl -pi -e 's/Hamt64//' hamt64-full.b

./summary.sh
