
echo "------------ wc ----------------"
time (cat test/84.txt | wc -w)
echo "------------ gowc ----------------"
time (cat test/84.txt | ./gowc)
echo "------------ test1 ----------------"
time (cat test/84.txt | ./test1)
echo "------------ test2 ----------------"
time (cat test/84.txt | ./test2)