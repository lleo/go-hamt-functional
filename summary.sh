#!/usr/bin/env bash

echo "Builtin Map" vs "Functional Hamt32 w/HybridTables"
benchcmp map.b hamt32-hybrid.b

echo
echo "Builtin Map" vs "Functional Hamt64 w/CompressedTables"
benchcmp map.b hamt64-comp.b

echo
echo "Functional Hamt32 w/HybridTables" vs "Functional Hamt32 w/CompressedTables"
benchcmp hamt32-hybrid.b hamt32-comp.b

echo
echo "Functional Hamt32 w/HybridTables" vs "Functional Hamt32 w/FullTables"
benchcmp hamt32-hybrid.b hamt32-full.b

echo
echo "Functional Hamt64 w/CompressedTables" vs "Functional Hamt64 w/FullTables"
benchcmp hamt64-comp.b hamt64-full.b
