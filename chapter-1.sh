#!/bin/bash
# let's print content like csv file in shell script using printf
#colored output of a shell script

echo -e "\e[1;33m === CSV FORMAT EXAMPLE ===\e[0m"

printf "%-5s %-10s %-4s\n" S.no Name Percentage
printf "%-5s %-10s %-4.1f\n" 1     Jack   45.456
printf "%-5s %-10s %-4.1f\n" 2     James   65.563     
printf "%-5s %-10s %-4.1f\n" 2     Jill   75.332
