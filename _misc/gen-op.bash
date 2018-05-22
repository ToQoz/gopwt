#!/bin/bash

set -e

(
  echo "// genereated by $(basename "$0")"
  echo "package translatedassert"
  echo

  uintegers=(uint uint8 uint16 uint32 uint64)
  integers=(uint uint8 uint16 uint32 uint64 int int8 int16 int32 int64)

  types=(uint uint8 uint16 uint32 uint64 int int8 int16 int32 int64 float32 float64 complex64 complex128)
  for t in "${types[@]}"
  do
    echo "func OpSUB_${t}(x $t, y $t) $t {
    return x - y
  }"
    echo
  done
  for t in "${types[@]}"
  do
    echo "func OpMUL_${t}(x $t, y $t) $t {
    return x * y
  }"
    echo
  done
  for t in "${types[@]}"
  do
    echo "func OpQUO_${t}(x $t, y $t) $t {
    return x / y
  }"
    echo
  done

  types=(string uint uint8 uint16 uint32 uint64 int int8 int16 int32 int64 float32 float64 complex64 complex128)
  for t in "${types[@]}"
  do
    echo "func OpADD_${t}(x $t, y $t) $t {
    return x + y
  }"
    echo
  done

  for t in "${integers[@]}"
  do
    echo "func OpREM_${t}(x $t, y $t) $t {
    return x % y
  }"
  done
  for t in "${integers[@]}"
  do
    echo "func OpAND_${t}(x $t, y $t) $t {
    return x & y
  }"
    echo
  done
  for t in "${integers[@]}"
  do
    echo "func OpOR_${t}(x $t, y $t) $t {
    return x | y
  }"
    echo
  done
  for t in "${integers[@]}"
  do
    echo "func OpXOR_${t}(x $t, y $t) $t {
    return x ^ y
  }"
    echo
  done
  for t in "${integers[@]}"
  do
    echo "func OpANDNOT_${t}(x $t, y $t) $t {
    return x &^ y
  }"
  done

  for xt in "${integers[@]}"
  do
    for yt in "${uintegers[@]}"
    do
      echo "func OpSHL_${xt}_${yt}(x $xt, y $yt) $xt {
    return x << y
  }"
      echo
    done
  done

  for xt in "${integers[@]}"
  do
    for yt in "${uintegers[@]}"
    do
      echo "func OpSHR_${xt}_${yt}(x $xt, y $yt) $xt {
    return x >> y
  }"
      echo
    done
  done

  echo "func OpLOR_bool(x bool, y bool) bool {
    return x || y
  }"
  echo
  echo "func OpLAND_bool(x bool, y bool) bool {
    return x && y
  }"
) > translatedassert/op.go
gofmt -w translatedassert/op.go

(
  echo "// genereated by $(basename "$0")"
  echo "package translatedassert"
  echo
  echo "var ops = map[string]interface{}{"
  grep 'func Op' translatedassert/op.go | awk '{print $2}' | awk -F'(' '{print $1}' | xargs -I{} echo '"{}": {},'
  echo "}"
)> translatedassert/op_map.go
gofmt -w translatedassert/op_map.go
