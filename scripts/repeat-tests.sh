counter=0
while [ $counter -lt 100 ]; do
  go test -v -count=1 -run ^TestGrantedLicenseAffectsCountsAndDetails$ authz/bootstrap
  result=$?
  if [ $result -ne 0 ]; then
    echo "Command failed with status $result after $counter runs."
    exit 1
  fi
  ((counter++))
done
echo "Command succeeded 100 times."
