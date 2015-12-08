#!/bin/sh
TARGET_FILES="app_rs232"
for FILE in ${TARGET_FILES}
do
  go install ${FILE}
done
