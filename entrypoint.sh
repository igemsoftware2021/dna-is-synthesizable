#!/bin/sh -l

echo $1 $2 $3 $4 $5 $6 $7 $8
eval "/annotator -i $1 -o $2 -r $3 -u $4 -p $5 -c $6 -s $7 -a $8"
