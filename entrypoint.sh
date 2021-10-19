#!/bin/sh -l
if [ -z  "$8" ] 
then  
  eval "/dna-is-synthesizable -i $1 -o $2 -r $3 -u $4 -p $5 -c $6 -s $7"
else
  eval "/dna-is-synthesizable -i $1 -o $2 -r $3 -u $4 -p $5 -c $6 -s $7 -a"
fi
