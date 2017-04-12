#!/usr/bin/env bash

for f in "$@"; do
	echo $f

	echo "perl -pi -e 's/32/64/g' $f"
	perl -pi -e 's/32/64/g' $f

	echo "perl -pi -e 's/30/60/g' $f"
	perl -pi -e 's/30/60/g' $f

	echo "perl -pi -e 's/\bsix\b/ten/g'"
	perl -pi -e 's/\bsix\b/ten/g' $f

	echo "perl -pi -e 's/\b5\b/6/g'"
	perl -pi -e 's/\b5\b/6/g' $f

done
