#!/bin/sh

set -e

echo "./wait && $@"

./wait && $@
